package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

type MigrationRunner struct {
	DB        *sql.DB
	Directory string
}

type Migration struct {
	Version int64
	Name    string
	Up      string
	Down    string
}

var migrationPattern = regexp.MustCompile(
	`^([0-9]+)_(.+)\.(up|down)\.sql$`,
)

func NewMigrationRunner(db *sql.DB, directory string) *MigrationRunner {
	return &MigrationRunner{
		DB:        db,
		Directory: directory,
	}
}

// Up applies all migrations that haven't been executed yet.
func (r *MigrationRunner) Up(ctx context.Context) error {
	if err := r.createMigrationTable(ctx); err != nil {
		return err
	}

	migrations, err := r.loadMigrations()
	if err != nil {
		return err
	}

	applied, err := r.appliedVersions(ctx)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if applied[migration.Version] {
			continue
		}

		if err := r.apply(ctx, migration); err != nil {
			return err
		}

		fmt.Printf("Applied migration %06d: %s\n",
			migration.Version, migration.Name)
	}

	return nil
}

// Down rolls back the latest applied migration.
func (r *MigrationRunner) Down(ctx context.Context) error {
	if err := r.createMigrationTable(ctx); err != nil {
		return err
	}

	var version int64
	err := r.DB.QueryRowContext(ctx,
		`SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`,
	).Scan(&version)

	if err == sql.ErrNoRows {
		fmt.Println("No migrations to roll back.")
		return nil
	}
	if err != nil {
		return fmt.Errorf("get latest migration: %w", err)
	}

	migrations, err := r.loadMigrations()
	if err != nil {
		return err
	}

	var migration *Migration
	for i := range migrations {
		if migrations[i].Version == version {
			migration = &migrations[i]
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration %d not found", version)
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, migration.Down); err != nil {
		return fmt.Errorf("execute down migration %d: %w", version, err)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM schema_migrations WHERE version = $1`, version,
	); err != nil {
		return fmt.Errorf("remove migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit rollback: %w", err)
	}

	fmt.Printf("Rolled back migration %06d: %s\n",
		migration.Version, migration.Name)

	return nil
}

func (r *MigrationRunner) createMigrationTable(ctx context.Context) error {
	_, err := r.DB.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}

	return nil
}

func (r *MigrationRunner) appliedVersions(ctx context.Context) (map[int64]bool, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT version FROM schema_migrations`,
	)
	if err != nil {
		return nil, fmt.Errorf("get applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int64]bool)

	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return applied, nil
}

func (r *MigrationRunner) apply(ctx context.Context, migration Migration) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, migration.Up); err != nil {
		return fmt.Errorf(
			"execute migration %d: %w",
			migration.Version, err,
		)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`,
		migration.Version, migration.Name,
	); err != nil {
		return fmt.Errorf("record migration %d: %w", migration.Version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %d: %w", migration.Version, err)
	}

	return nil
}

func (r *MigrationRunner) loadMigrations() ([]Migration, error) {
	entries, err := os.ReadDir(r.Directory)
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	migrations := make(map[int64]*Migration)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if filepath.Ext(filename) != ".sql" {
			continue
		}

		matches := migrationPattern.FindStringSubmatch(filename)
		if matches == nil {
			return nil, fmt.Errorf("invalid migration filename: %s", filename)
		}

		version, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid migration version: %s", filename)
		}

		name := matches[2]
		direction := matches[3]

		if migrations[version] == nil {
			migrations[version] = &Migration{
				Version: version,
				Name:    name,
			}
		}

		migration := migrations[version]

		if migration.Name != name {
			return nil, fmt.Errorf("inconsistent migration name for version %d", version)
		}

		path := filepath.Join(r.Directory, filename)
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", filename, err)
		}

		if direction == "up" {
			if migration.Up != "" {
				return nil, fmt.Errorf("duplicate up migration: %s", filename)
			}
			migration.Up = string(content)
		} else {
			if migration.Down != "" {
				return nil, fmt.Errorf("duplicate down migration: %s", filename)
			}
			migration.Down = string(content)
		}
	}

	result := make([]Migration, 0, len(migrations))

	for _, migration := range migrations {
		if migration.Up == "" || migration.Down == "" {
			return nil, fmt.Errorf(
				"migration %d must have both up and down files",
				migration.Version,
			)
		}
		result = append(result, *migration)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})

	return result, nil
}

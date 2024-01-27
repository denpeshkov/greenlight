package postgres

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"slices"

	"github.com/denpeshkov/greenlight/internal/multierr"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// MigrationType is type of migration to run: UP or DOWN
type MigrationType uint8

// String returns up or down depending on migration type.
func (mt MigrationType) String() string {
	if mt == UP {
		return "up"
	}
	return "down"
}

const (
	UP MigrationType = iota + 1
	DOWN
)

// FIXME execute migrations in a DB transaction.

// Migrate looks at the currently active migration version and applies all up or down migrations, depending on the provided argument.
func (db *DB) Migrate(t MigrationType) error {
	op := "postgres.DB.Migrate"

	db.logger.Debug("start database migration", "type", t.String())

	// Creates a migrations table.
	// Table is based on: https://github.com/golang-migrate/migrate.
	// version 0 means that no migrations are applied (all rollback-ed)
	query := `CREATE TABLE IF NOT EXISTS migrations (version bigint PRIMARY KEY DEFAULT 0, dirty boolean NOT NULL DEFAULT FALSE)`
	if _, err := db.db.Exec(query); err != nil {
		return fmt.Errorf("%s: create migration table: %w", op, err)
	}

	version, dirty, err := db.migrationState()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if dirty {
		return fmt.Errorf("%s: current DB has dirty version=%d", op, version)
	}

	names, err := read(t)
	if err != nil {
		return fmt.Errorf("%s: get migration files: %w", op, err)
	}

	slices.Sort(names)
	if t == UP {
		version, err = db.migrateUp(names, version)
	} else {
		version, err = db.migrateDown(names, version)
	}
	if err != nil {
		dirty = true
		err = fmt.Errorf("%s: DB left with dirty version=%d: %w", op, version, err)
	}

	// Update the migration table even in the presence of an error; version will be of the failed migration and dirty will be true.
	if _, err2 := db.db.Exec(`TRUNCATE TABLE migrations`); err2 != nil {
		return fmt.Errorf("%s: update migration table: %w", op, multierr.Join(err2, err))
	}
	if _, err2 := db.db.Exec(`INSERT INTO migrations (version, dirty) VALUES ($1, $2)`, version, dirty); err2 != nil {
		return fmt.Errorf("%s: update migration table: %w", op, multierr.Join(err2, err))
	}
	return err
}

func (db *DB) migrateUp(names []string, version int) (int, error) {
	op := "postgres.DB.migrateUp"

	for version < len(names) {
		name := names[version]
		if err := db.migrateFile(name); err != nil {
			return version, fmt.Errorf("%s: migration file %q: %w", op, name, err)
		}
		db.logger.Debug("database migration", "migration_type", "UP", "file", name)
		version++
	}
	return version, nil
}

func (db *DB) migrateDown(names []string, version int) (int, error) {
	op := "postgres.DB.migrateDown"

	for version >= 1 {
		name := names[version-1]
		if err := db.migrateFile(name); err != nil {
			return version, fmt.Errorf("%s: migration file %q: %w", op, name, err)
		}
		db.logger.Debug("database migration", "migration_type", "DOWN", "file", name)
		version--
	}
	return version, nil
}

// migrate runs a single migration file.
func (db *DB) migrateFile(name string) error {
	op := "postgres.DB.migrateFile"

	// Read and execute migration file.
	if buf, err := migrationFS.ReadFile(name); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	} else if _, err = db.db.Exec(string(buf)); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// migrationState returns current migration version and dirty state.
// In case of an error returned version is 0 and dirty is false.
func (db *DB) migrationState() (version int, dirty bool, err error) {
	op := "postgres.DB.migrationState"

	if err = db.db.QueryRow(`SELECT version, dirty FROM migrations`).Scan(&version, &dirty); err != nil && err != sql.ErrNoRows {
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}
	return version, dirty, nil
}

// read returns the names of all up or down migration files.
func read(t MigrationType) ([]string, error) {
	op := "postgres.read"

	pattern := fmt.Sprintf("migrations/*.%s.sql", t.String())
	names, err := fs.Glob(migrationFS, pattern)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return names, nil
}

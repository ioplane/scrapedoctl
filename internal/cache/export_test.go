package cache

import "database/sql"

// Exported for testing.
var (
	ExpandPath         = expandPath
	NormalizeAndHashFn = NormalizeAndHash
)

func (s *Store) Database() *sql.DB {
	return s.database
}

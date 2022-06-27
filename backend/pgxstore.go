package backend

import (
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/yi-jiayu/pgxstore"
)

type Store interface {
	sessions.Store
}

type store struct {
	*pgxstore.PGStore
}

var _ Store = new(store)

func NewStore(db *pgxpool.Pool, keyPairs ...[]byte) (Store, error) {
	p, err := pgxstore.NewPGStoreFromConn(db, keyPairs...)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgstore from pool: %w", err)
	}

	p.MaxLength(0)

	return &store{p}, nil
}

func (s *store) Options(options sessions.Options) {
	s.PGStore.Options = options.ToGorillaOptions()
}

func (s *store) MaxLength(l int) {
	s.PGStore.MaxLength(l)
}

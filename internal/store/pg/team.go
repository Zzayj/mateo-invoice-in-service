package pg

import (
	"context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func (s *Store) GetBoostedTeamIds(ctx context.Context) ([]string, error) {
	const query = `
		SELECT id
		FROM "Team"
		WHERE is_boosted = true
	`
	rows, err := s.conn.Query(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("failed to get boosted team ids")
		return nil, errors.Wrap(err, "failed to get boosted team ids")
	}
	defer rows.Close()

	var teamIds []string
	for rows.Next() {
		var teamId string
		if err := rows.Scan(&teamId); err != nil {
			log.Error().Err(err).Msg("failed to get boosted team ids")
			return nil, errors.Wrap(err, "failed to get boosted team ids")
		}
		teamIds = append(teamIds, teamId)
	}

	return teamIds, nil
}

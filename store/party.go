package store

import "context"

type Party struct {
	Name string
	ID   int
}

const getPartiesQuery = `select id_party, name from parties`

func (p *PGStore) GetParties(ctx context.Context) ([]Party, error) {
	rows, err := p.db.Query(ctx, getPartiesQuery)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var parties []Party

	for rows.Next() {
		var party Party
		err := rows.Scan(&party.ID, &party.Name)
		if err != nil {
			return nil, err
		}
		parties = append(parties, party)
	}
	return parties, nil
}

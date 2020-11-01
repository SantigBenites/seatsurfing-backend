package main

import (
	"strings"
	"sync"
)

type LocationRepository struct {
}

type Location struct {
	ID             string
	OrganizationID string
	Name           string
	MapWidth       uint
	MapHeight      uint
	MapMimeType    string
}

type LocationMap struct {
	MimeType string
	Width    uint
	Height   uint
	Data     []byte
}

var locationRepository *LocationRepository
var locationRepositoryOnce sync.Once

func GetLocationRepository() *LocationRepository {
	locationRepositoryOnce.Do(func() {
		locationRepository = &LocationRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS locations (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"organization_id uuid NOT NULL, " +
			"name VARCHAR NOT NULL, " +
			"map_mimetype VARCHAR DEFAULT ''," +
			"map_data BYTEA," +
			"map_width INTEGER DEFAULT 0," +
			"map_height INTEGER DEFAULT 0," +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
	})
	return locationRepository
}

func (r *LocationRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	// No updates yet
}

func (r *LocationRepository) Create(e *Location) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO locations "+
		"(organization_id, name) "+
		"VALUES ($1, $2) "+
		"RETURNING id",
		e.OrganizationID, e.Name).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *LocationRepository) GetOne(id string) (*Location, error) {
	e := &Location{}
	err := GetDatabase().DB().QueryRow("SELECT id, organization_id, name, map_mimetype, map_width, map_height "+
		"FROM locations "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.OrganizationID, &e.Name, &e.MapMimeType, &e.MapWidth, &e.MapHeight)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *LocationRepository) GetByKeyword(organizationID string, keyword string) ([]*Location, error) {
	var result []*Location
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, name, map_mimetype, map_width, map_height "+
		"FROM locations "+
		"WHERE organization_id = $1 AND LOWER(name) LIKE '%' || $2 || '%' "+
		"ORDER BY name", organizationID, strings.ToLower(keyword))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &Location{}
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.Name, &e.MapMimeType, &e.MapWidth, &e.MapHeight)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *LocationRepository) GetAll(organizationID string) ([]*Location, error) {
	var result []*Location
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, name, map_mimetype, map_width, map_height "+
		"FROM locations "+
		"WHERE organization_id = $1 "+
		"ORDER BY name", organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &Location{}
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.Name, &e.MapMimeType, &e.MapWidth, &e.MapHeight)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *LocationRepository) Update(e *Location) error {
	_, err := GetDatabase().DB().Exec("UPDATE locations SET "+
		"organization_id = $1, "+
		"name = $2 "+
		"WHERE id = $3",
		e.OrganizationID, e.Name, e.ID)
	return err
}

func (r *LocationRepository) Delete(e *Location) error {
	if _, err := GetDatabase().DB().Exec("DELETE FROM bookings WHERE bookings.space_id IN (SELECT spaces.id FROM spaces WHERE spaces.location_id = $1)", e.ID); err != nil {
		return err
	}
	if _, err := GetDatabase().DB().Exec("DELETE FROM spaces WHERE location_id = $1", e.ID); err != nil {
		return err
	}
	_, err := GetDatabase().DB().Exec("DELETE FROM locations WHERE id = $1", e.ID)
	return err
}

func (r *LocationRepository) DeleteAll(organizationID string) error {
	if _, err := GetDatabase().DB().Exec("DELETE FROM bookings WHERE "+
		"bookings.space_id IN (SELECT spaces.id FROM spaces WHERE "+
		"spaces.location_id IN (SELECT locations.id FROM locations WHERE locations.organization_id = $1)"+
		")", organizationID); err != nil {
		return err
	}
	if _, err := GetDatabase().DB().Exec("DELETE FROM spaces WHERE spaces.location_id IN (SELECT locations.id FROM locations WHERE locations.organization_id = $1)", organizationID); err != nil {
		return err
	}
	_, err := GetDatabase().DB().Exec("DELETE FROM locations WHERE organization_id = $1", organizationID)
	return err
}

func (r *LocationRepository) GetCount(organizationID string) (int, error) {
	var res int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(id) "+
		"FROM locations "+
		"WHERE organization_id = $1",
		organizationID).Scan(&res)
	return res, err
}

func (r *LocationRepository) SetMap(e *Location, locationMap *LocationMap) error {
	_, err := GetDatabase().DB().Exec("UPDATE locations SET "+
		"map_mimetype = $1, "+
		"map_data = $2, "+
		"map_width = $3, "+
		"map_height = $4 "+
		"WHERE id = $5",
		locationMap.MimeType, locationMap.Data, locationMap.Width, locationMap.Height, e.ID)
	return err
}

func (r *LocationRepository) GetMap(location *Location) (*LocationMap, error) {
	e := &LocationMap{}
	err := GetDatabase().DB().QueryRow("SELECT map_mimetype, map_data, map_width, map_height "+
		"FROM locations "+
		"WHERE id = $1",
		location.ID).Scan(&e.MimeType, &e.Data, &e.Width, &e.Height)
	if err != nil {
		return nil, err
	}
	return e, nil
}
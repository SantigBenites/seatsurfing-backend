package main

import (
	"math"
	"sync"
	"time"
)

type BookingRepository struct {
}

type Booking struct {
	ID      string
	UserID  string
	SpaceID string
	Enter   time.Time
	Leave   time.Time
}

type BookingDetails struct {
	Space     SpaceDetails
	UserEmail string
	Booking
}

var bookingRepository *BookingRepository
var bookingRepositoryOnce sync.Once

func GetBookingRepository() *BookingRepository {
	bookingRepositoryOnce.Do(func() {
		bookingRepository = &BookingRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS bookings (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"user_id uuid NOT NULL, " +
			"space_id uuid NOT NULL, " +
			"enter_time TIMESTAMP NOT NULL, " +
			"leave_time TIMESTAMP NOT NULL, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON bookings(user_id)")
		if err != nil {
			panic(err)
		}
	})
	return bookingRepository
}

func (r *BookingRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	// No updates yet
}

func (r *BookingRepository) Create(e *Booking) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO bookings "+
		"(user_id, space_id, enter_time, leave_time) "+
		"VALUES ($1, $2, $3, $4) "+
		"RETURNING id",
		e.UserID, e.SpaceID, e.Enter, e.Leave).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *BookingRepository) GetOne(id string) (*BookingDetails, error) {
	e := &BookingDetails{}
	err := GetDatabase().DB().QueryRow("SELECT bookings.id, bookings.user_id, bookings.space_id, bookings.enter_time, bookings.leave_time, "+
		"spaces.id, spaces.location_id, spaces.name, "+
		"locations.id, locations.organization_id, locations.name, "+
		"users.email "+
		"FROM bookings "+
		"INNER JOIN spaces ON bookings.space_id = spaces.id "+
		"INNER JOIN locations ON spaces.location_id = locations.id "+
		"INNER JOIN users ON bookings.user_id = users.id "+
		"WHERE bookings.id = $1",
		id).Scan(&e.ID, &e.UserID, &e.SpaceID, &e.Enter, &e.Leave, &e.Space.ID, &e.Space.LocationID, &e.Space.Name, &e.Space.Location.ID, &e.Space.Location.OrganizationID, &e.Space.Location.Name, &e.UserEmail)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *BookingRepository) GetAllByOrg(organizationID string, startTime, endTime time.Time) ([]*BookingDetails, error) {
	var result []*BookingDetails
	rows, err := GetDatabase().DB().Query("SELECT bookings.id, bookings.user_id, bookings.space_id, bookings.enter_time, bookings.leave_time, "+
		"spaces.id, spaces.location_id, spaces.name, "+
		"locations.id, locations.organization_id, locations.name, "+
		"users.email "+
		"FROM bookings "+
		"INNER JOIN spaces ON bookings.space_id = spaces.id "+
		"INNER JOIN locations ON spaces.location_id = locations.id "+
		"INNER JOIN users ON bookings.user_id = users.id "+
		"WHERE locations.organization_id = $1 AND leave_time >= $2 AND enter_time <= $3 "+
		"ORDER BY enter_time", organizationID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &BookingDetails{}
		err = rows.Scan(&e.ID, &e.UserID, &e.SpaceID, &e.Enter, &e.Leave, &e.Space.ID, &e.Space.LocationID, &e.Space.Name, &e.Space.Location.ID, &e.Space.Location.OrganizationID, &e.Space.Location.Name, &e.UserEmail)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *BookingRepository) GetAllByUser(userID string, startTime time.Time) ([]*BookingDetails, error) {
	var result []*BookingDetails
	rows, err := GetDatabase().DB().Query("SELECT bookings.id, bookings.user_id, bookings.space_id, bookings.enter_time, bookings.leave_time, "+
		"spaces.id, spaces.location_id, spaces.name, "+
		"locations.id, locations.organization_id, locations.name, "+
		"users.email "+
		"FROM bookings "+
		"INNER JOIN spaces ON bookings.space_id = spaces.id "+
		"INNER JOIN locations ON spaces.location_id = locations.id "+
		"INNER JOIN users ON bookings.user_id = users.id "+
		"WHERE user_id = $1 AND leave_time >= $2 "+
		"ORDER BY enter_time", userID, startTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &BookingDetails{}
		err = rows.Scan(&e.ID, &e.UserID, &e.SpaceID, &e.Enter, &e.Leave, &e.Space.ID, &e.Space.LocationID, &e.Space.Name, &e.Space.Location.ID, &e.Space.Location.OrganizationID, &e.Space.Location.Name, &e.UserEmail)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}
func (r *BookingRepository) Update(e *Booking) error {
	_, err := GetDatabase().DB().Exec("UPDATE bookings SET "+
		"user_id = $1, "+
		"space_id = $2, "+
		"enter_time = $3, "+
		"leave_time = $4 "+
		"WHERE id = $5",
		e.UserID, e.SpaceID, e.Enter, e.Leave, e.ID)
	return err
}

func (r *BookingRepository) Delete(e *BookingDetails) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM bookings WHERE id = $1", e.ID)
	return err
}

func (r *BookingRepository) GetCount(organizationID string) (int, error) {
	var res int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(bookings.id) "+
		"FROM bookings "+
		"INNER JOIN spaces ON spaces.id = bookings.space_id "+
		"INNER JOIN locations ON locations.id = spaces.location_id "+
		"WHERE locations.organization_id = $1",
		organizationID).Scan(&res)
	return res, err
}

func (r *BookingRepository) GetCountDateRange(organizationID string, enter, leave time.Time) (int, error) {
	var res int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(bookings.id) "+
		"FROM bookings "+
		"INNER JOIN spaces ON spaces.id = bookings.space_id "+
		"INNER JOIN locations ON locations.id = spaces.location_id "+
		"WHERE locations.organization_id = $1 AND ("+
		"($2 BETWEEN enter_time AND leave_time) OR "+
		"($3 BETWEEN enter_time AND leave_time) OR "+
		"(enter_time BETWEEN $2 AND $3) OR "+
		"(leave_time BETWEEN $2 AND $3)"+
		")",
		organizationID, enter, leave).Scan(&res)
	return res, err
}

func (r *BookingRepository) GetTotalBookedMinutes(organizationID string, enter, leave time.Time) (int, error) {
	var totalBookedMinutes float64
	err := GetDatabase().DB().QueryRow("SELECT SUM(EXTRACT(EPOCH FROM (LEAST(leave_time, $3) - GREATEST(enter_time, $2)))/60) "+
		"FROM bookings "+
		"INNER JOIN spaces ON spaces.id = bookings.space_id "+
		"INNER JOIN locations ON locations.id = spaces.location_id "+
		"WHERE locations.organization_id = $1 AND ("+
		"($2 BETWEEN enter_time AND leave_time) OR "+
		"($3 BETWEEN enter_time AND leave_time) OR "+
		"(enter_time BETWEEN $2 AND $3) OR "+
		"(leave_time BETWEEN $2 AND $3)"+
		")",
		organizationID, enter, leave).Scan(&totalBookedMinutes)
	return int(math.RoundToEven(totalBookedMinutes)), err
}

func (r *BookingRepository) GetLoad(organizationID string, enter, leave time.Time) (int, error) {
	totalBookedMinutes, err := r.GetTotalBookedMinutes(organizationID, enter, leave)
	if err != nil {
		return 0, err
	}
	numSpaces, err := GetSpaceRepository().GetCount(organizationID)
	if err != nil {
		return 0, err
	}
	totalTimeMinutes := leave.Sub(enter).Minutes() * float64(numSpaces)
	res := float64(totalBookedMinutes) / float64(totalTimeMinutes) * float64(100)
	if res > 100.0 {
		res = 100.0
	}
	return int(math.RoundToEven(res)), nil
}

func (r *BookingRepository) GetConflicts(spaceID string, enter time.Time, leave time.Time, excludeBookingID string) ([]*Booking, error) {
	var result []*Booking
	rows, err := GetDatabase().DB().Query("SELECT id, user_id, space_id, enter_time, leave_time "+
		"FROM bookings "+
		"WHERE id::text != $1 AND space_id = $2 AND ("+
		"($3 BETWEEN enter_time AND leave_time) OR "+
		"($4 BETWEEN enter_time AND leave_time) OR "+
		"(enter_time BETWEEN $3 AND $4) OR "+
		"(leave_time BETWEEN $3 AND $4)"+
		") "+
		"ORDER BY enter_time", excludeBookingID, spaceID, enter, leave)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &Booking{}
		err = rows.Scan(&e.ID, &e.UserID, &e.SpaceID, &e.Enter, &e.Leave)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}
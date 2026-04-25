package domain

// Status represents the operational state of a driver.
type Status string

const (
	StatusOffline Status = "offline"
	StatusOnline  Status = "online"
	StatusOnTrip  Status = "on_trip"
)

// CanTransitionTo validates the state machine rules:
//   offline → online
//   online  → offline | on_trip
//   on_trip → online
func (s Status) CanTransitionTo(next Status) bool {
	allowed := map[Status][]Status{
		StatusOffline: {StatusOnline},
		StatusOnline:  {StatusOffline, StatusOnTrip},
		StatusOnTrip:  {StatusOnline},
	}
	for _, a := range allowed[s] {
		if a == next {
			return true
		}
	}
	return false
}

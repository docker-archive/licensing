package licensing

type State string

const (
	Active    State = "active"
	Expired   State = "expired"
	Cancelled State = "cancelled"
	Preparing State = "preparing"
	Failed    State = "failed"
)

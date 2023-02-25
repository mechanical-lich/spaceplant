package components

// DecayingComponent Component base component interface
type DecayingComponent interface {
	Decay() bool
	GetType() string
}

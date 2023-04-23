package servicediscovery

type Discovery interface {
	PickOne() string
	PickAll() []string
	Watch()
	Exit()
}

type Register interface {
	Run()
	Exit()
}

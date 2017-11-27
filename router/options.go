package router

type RouterOptions struct {
	// Port is the HTTP port the router will listen on
	Port     int
	Hostname string
}

var defaultOptions = RouterOptions{
	Port:     8080,
	Hostname: "localhost",
}

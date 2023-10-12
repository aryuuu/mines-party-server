package configs

import "os"

type service struct {
	Port        string
	ServiceName string
}

func initService() *service {
	result := &service{
		Port:        os.Getenv("PORT"),
		ServiceName: os.Getenv("SERVICE_NAME"),
	}

	return result
}

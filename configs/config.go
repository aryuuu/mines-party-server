package configs

// Service :nodoc:
var Service *service

// Constant :nodoc:
var Constant *constant

func init() {
	Service = initService()
	Constant = initConstant()
}

package kpture

type InvalidEnvParamError struct {
	param string
}

func (e InvalidEnvParamError) Error() string {
	return "invalid environment parameter: " + e.param
}

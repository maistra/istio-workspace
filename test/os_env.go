package test

import "os"

type EnvVars struct {
	originalEnvs map[string]string
}

func TemporaryEnvVars() *EnvVars {
	return &EnvVars{
		originalEnvs: map[string]string{},
	}
}

func (env *EnvVars) SetAll(envVars map[string]string) {
	for k, v := range envVars {
		env.Set(k, v)
	}
}

func (env *EnvVars) Set(key, value string) {
	env.originalEnvs[key] = os.Getenv(key)
	if value != "" {
		_ = os.Setenv(key, value)
	} else {
		_ = os.Unsetenv(key)
	}
}

func (env *EnvVars) Restore() {
	for k, v := range env.originalEnvs {
		if v != "" {
			_ = os.Setenv(k, v)
		} else {
			_ = os.Unsetenv(k)
		}
	}
}

package test

import "os"

func TemporaryEnvVars(keyValues ...string) func() {
	if len(keyValues)%2 != 0 {
		panic("you should supply even amount of key-value arguments")
	}

	vars := map[string]string{}
	for i := 0; i < len(keyValues); i += 2 {
		vars[keyValues[i]] = keyValues[i+1]
	}
	originalEnvs := map[string]string{}

	for key, value := range vars {
		originalEnvs[key] = os.Getenv(key)
		if value != "" {
			_ = os.Setenv(key, value)
		} else {
			_ = os.Unsetenv(key)
		}
	}

	return func() {
		for k, v := range originalEnvs {
			if v != "" {
				_ = os.Setenv(k, v)
			} else {
				_ = os.Unsetenv(k)
			}
		}
	}
}

func TemporaryUnsetEnvVars(keys ...string) func() { //nolint:unused //reason linter is drunk
	originalEnvs := map[string]string{}

	for _, key := range keys {
		originalEnvs[key] = os.Getenv(key)
		_ = os.Unsetenv(key)
	}

	return func() {
		for k, v := range originalEnvs {
			if v != "" {
				_ = os.Setenv(k, v)
			} else {
				_ = os.Unsetenv(k)
			}
		}
	}
}

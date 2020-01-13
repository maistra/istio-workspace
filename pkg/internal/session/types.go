package session

import "fmt"

// DeploymentNotFoundError denotes failing to find deployment.
type DeploymentNotFoundError struct {
	name string
}

// Error returns the formatted deployment error.
func (dnfe DeploymentNotFoundError) Error() string {
	return fmt.Sprintf("no Deployment or DeploymentConfig found for target '%s'", dnfe.name)
}

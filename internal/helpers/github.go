package helpers

import (
	"fmt"
	"strings"
)

func SplitOwnerRepo(nameWithOwner string) (name string, owner string, err error) {
	parts := strings.Split(nameWithOwner, "/")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid nameWithOwner: %s", nameWithOwner)
	}

	return parts[0], parts[1], nil
}

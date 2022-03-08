package config

import "fmt"

func ExamplePasswordHash() {
	seed := []byte("myseed")
	hash := PasswordHashFromSeed(seed, []byte("mypass"))
	fmt.Println(JoinSeedAndHash(seed, hash))

	// Output:
	// bXlzZWVk:HMSxrg1cYphaPuUYUbtbl/htep/tVYYIQAuvkNMVpw0
}

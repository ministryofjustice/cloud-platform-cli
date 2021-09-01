package kubecfg

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/golang-jwt/jwt"
)

// ShowGithubTeams prints tokens claims that are useful (among other things) for cloud platform
// users to identify which github teams they belongs to.
func ShowGithubTeams(f string) error {
	tokenString, err := getToken(f)
	if err != nil {
		return err
	}

	for _, t := range tokenString {
		claims, err := getTokenClaims(t)
		if err != nil {
			return err
		}
		fmt.Println(prettyPrint(claims))
	}

	return nil
}

func prettyPrint(i *jwt.MapClaims) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func getToken(f string) ([]string, error) {
	var tokens []string

	file, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := regexp.MustCompile(`id-token:\s(.*)`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		if r.MatchString(scanner.Text()) {
			match := r.FindStringSubmatch(scanner.Text())
			tokens = append(tokens, match[1])
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	if len(tokens) == 0 {
		return nil, errors.New("kubeconfig file DOES NOT contain id-token")
	}

	return tokens, nil
}

func getTokenClaims(t string) (*jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	token, _, err := new(jwt.Parser).ParseUnverified(t, claims)
	if err != nil {
		return nil, err
	}

	if _, ok := token.Claims.(jwt.MapClaims); !ok {
		return nil, err
	}

	return &claims, nil
}

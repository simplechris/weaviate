//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2021 SeMI Technologies B.V. All rights reserved.
//
//  CONTACT: hello@semi.technology
//

package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// FromEnv takes a *Config as it will respect initial config that has been
// provided by other means (e.g. a config file) and will only extend those that
// are set
func FromEnv(config *Config) error {
	if enabled(os.Getenv("AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED")) {
		config.Authentication.AnonymousAccess.Enabled = true
	}

	if enabled(os.Getenv("AUTHENTICATION_OIDC_ENABLED")) {
		config.Authentication.OIDC.Enabled = true

		if enabled(os.Getenv("AUTHENTICATION_OIDC_SKIP_CLIENT_ID_CHECK")) {
			config.Authentication.OIDC.SkipClientIDCheck = true
		}

		if v := os.Getenv("AUTHENTICATION_OIDC_ISSUER"); v != "" {
			config.Authentication.OIDC.Issuer = v
		}

		if v := os.Getenv("AUTHENTICATION_OIDC_CLIENT_ID"); v != "" {
			config.Authentication.OIDC.ClientID = v
		}

		if v := os.Getenv("AUTHENTICATION_OIDC_USERNAME_CLAIM"); v != "" {
			config.Authentication.OIDC.UsernameClaim = v
		}

		if v := os.Getenv("AUTHENTICATION_OIDC_GROUPS_CLAIM"); v != "" {
			config.Authentication.OIDC.GroupsClaim = v
		}
	}

	if enabled(os.Getenv("AUTHORIZATION_ADMINLIST_ENABLED")) {
		config.Authorization.AdminList.Enabled = true

		users := strings.Split(os.Getenv("AUTHORIZATION_ADMINLIST_USERS"), ",")
		roUsers := strings.Split(os.Getenv("AUTHORIZATION_ADMINLIST_READONLY_USERS"),
			",")

		config.Authorization.AdminList.ReadOnlyUsers = roUsers
		config.Authorization.AdminList.Users = users
	}

	if v := os.Getenv("PERSISTENCE_DATA_PATH"); v != "" {
		config.Persistence.DataPath = v
	}

	if v := os.Getenv("ORIGIN"); v != "" {
		config.Origin = v
	}

	if v := os.Getenv("CONTEXTIONARY_URL"); v != "" {
		config.Contextionary.URL = v
	}

	if v := os.Getenv("QUERY_DEFAULTS_LIMIT"); v != "" {
		asInt, err := strconv.Atoi(v)
		if err != nil {
			return errors.Wrapf(err, "parse QUERY_DEFAULTS_LIMIT as int")
		}

		config.QueryDefaults.Limit = int64(asInt)
	}

	if v := os.Getenv("DEFAULT_VECTORIZER_MODULE"); v != "" {
		config.DefaultVectorizerModule = v
	} else {
		// env not set, this could either mean, we already have a value from a file
		// or we explicitly want to set the value to "none"
		if config.DefaultVectorizerModule == "" {
			config.DefaultVectorizerModule = VectorizerModuleNone
		}
	}

	if v := os.Getenv("ENABLE_MODULES"); v != "" {
		config.EnableModules = v
	}

	return nil
}

const VectorizerModuleNone = "none"

// TODO: This should be retrieved dynamically from all installed modules
const VectorizerModuleText2VecContextionary = "text2vec-contextionary"

func enabled(value string) bool {
	if value == "" {
		return false
	}

	if value == "on" ||
		value == "enabeld" ||
		value == "1" ||
		value == "true" {
		return true
	}

	return false
}

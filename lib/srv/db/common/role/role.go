/*
Copyright 2021 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package role

import (
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/defaults"
	"github.com/gravitational/teleport/lib/services"
)

// DatabaseRoleMatchers returns role matchers based on the database.
func DatabaseRoleMatchers(db types.Database, user, database string) services.RoleMatchers {
	roleMatchers := services.RoleMatchers{
		services.NewDatabaseUserMatcher(db, user),
	}

	if matcher := databaseNameMatcher(db.GetProtocol(), database); matcher != nil {
		roleMatchers = append(roleMatchers, matcher)
	}
	return roleMatchers
}

// RequireDatabaseUserMatcher returns true if databases with provided protocol
// require database users.
func RequireDatabaseUserMatcher(protocol string) bool {
	return true // Always required.
}

// RequireDatabaseNameMatcher returns true if databases with provided protocol
// require database names.
func RequireDatabaseNameMatcher(protocol string) bool {
	return databaseNameMatcher(protocol, "") != nil
}

func databaseNameMatcher(dbProtocol, database string) *services.DatabaseNameMatcher {
	switch dbProtocol {
	case
		// In MySQL, unlike Postgres, "database" and "schema" are the same thing
		// and there's no good way to prevent users from performing cross-database
		// queries once they're connected, apart from granting proper privileges
		// in MySQL itself.
		//
		// As such, checking db_names for MySQL is quite pointless, so we only
		// check db_users. In the future, if we implement some sort of access controls
		// on queries, we might be able to restrict db_names as well e.g. by
		// detecting full-qualified table names like db.table, until then the
		// proper way is to use MySQL grants system.
		defaults.ProtocolMySQL,
		// Cockroach uses the same wire protocol as Postgres but handling of
		// databases is different and there's no way to prevent cross-database
		// queries so only apply RBAC to db_users.
		defaults.ProtocolCockroachDB,
		// Redis integration doesn't support schema access control.
		defaults.ProtocolRedis,
		// Cassandra integration doesn't support schema access control.
		defaults.ProtocolCassandra,
		// Elasticsearch integration doesn't support schema access control.
		defaults.ProtocolElasticsearch,
		// DynamoDB integration doesn't support schema access control.
		defaults.ProtocolDynamoDB,
		// Oracle integration doesn't support schema access control.
		defaults.ProtocolOracle:
		return nil
	default:
		return &services.DatabaseNameMatcher{Name: database}
	}
}

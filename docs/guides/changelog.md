---
page_title: "Changelog"
subcategory: ""
description: |-
  Changelog
---

# Changelog

## 5.4.1

**Fixes:**

* Prevent replacement of `azuresql_user` on import when `entraid_identifier` is set.

## 5.4.0

**Features:**

* Add support for Synapse dedicated via the `synapseserver` datasource.

**Changed:**

* Updates queries checking for resource existence to be compatible with both Synapse server and all other supported azuresql servers.

## 5.3.1

**Fixes:**

* Check cached connections are healthy before reusing them.
* `azuresql_login` check that a connection was obtained successfully before using it.

## 5.3.0

**Features:**

* `azuresql_user` now supports creating database users with a password.

## 5.2.1

**Bugfixes:**

* Fix view definition parsing to allow whitespace chacacters beyond spaces.

## 5.2.0

**Features:**

* Allow specifying the password in `azuresql_login`

## 5.1.0

**Bugfixes:**

* In `azuresql_user`, renamed the experimental `object_id` parameter to `entraid_identifier` to reflect that application users must be created using their client ID (Entra ID identifier) instead of their Object ID.

**Documentation:**

* Add page on CICD authentication.

## 5.0.1

**Features:**
* Add getting started guides

~> I made a mistake causing the version to suddenly jump from 0.4.2 to 5.0.0. 

## 0.5.0

**Features:**
* Add `azuresql_procedure` resource to manage stored procedures in SQL
* Add support for Fabric workspaces

## 0.4.2

**Features:**
* `azuresql_permission` can now grant permissions on `azuresql_function` resources.

**Bugfixes:**
* Fix: Login error caused by caching faulty connections
* Fix: Login error when using Environmental credentials
* Fix: Provider documentation - default value for `check_database_exists` is true
* Update go dependencies
  
## 0.4.1

**Features:**

**Bugfixes:**
* Fix: Error when renaming resource azuresql_database 
* Fix: Error when using upper case letters when creating an azuresql_function resource
* Fix: Bug in security_predicate continously detecting changes when the rule contains spaces
* Update provider dependencies

**Documentation:**
* Improve indentation docs

## 0.4.0

**Features:**
* Add `azuresql_view` as a data source
* Add `azuresql_database` as a resource
* Add optional parameters `subscription_id`, `check_server_exists` and `check_database_exists` which can be used to check database/server existence before connecting.

**Bugfixes:**
* Type conversion error when creating an `azuresql_function` using the raw api
* Exponential retry until Synapse pool has warmed up

**Documentation:**
* Start Changelog
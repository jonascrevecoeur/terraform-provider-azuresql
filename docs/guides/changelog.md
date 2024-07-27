---
page_title: "Changelog"
subcategory: ""
description: |-
  Changelog
---

# Changelog

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
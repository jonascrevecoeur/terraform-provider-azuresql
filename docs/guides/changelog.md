---
page_title: "Changelog"
subcategory: ""
description: |-
  Changelog
---

# Changelog

## 0.4.1

**Features:**

**Bugfixes:**
* Fix: Error when renaming resource azuresql_dabase 
* Fix: Error when using upper case letters when creating an azuresql_function resource

**Documentation:**
* Improve documentation formatting

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
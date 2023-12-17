---
page_title: "Changelog"
subcategory: ""
description: |-
  Changelog
---

# Changelog

## 0.4.0

**Features:**
* Add `azuresql_view` as a data source
* Add `azuresql_database` as a resource
* Add optional parameters `subscription_id`, `check_server_exists` and `check_database_exists` which can be used to check database/server existence before connecting.

**Bugfixes:**
* Type conversion error when creating an `azuresql_function` using the raw api

**Documentation:**
* Start Changelog
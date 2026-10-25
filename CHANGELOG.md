# Changelog

## [v4.2.0](https://github.com/contentways/poweradmin-go/releases/tag/v4.2.0)

### Features

- **dynamic-dns**: add a client for POST /v2/dynamic-dns
- **groups**: confirm deleting a group that still owns zones
- **users**: assign groups on creation and read memberships
- **zones**: set the SOA-EDIT-API policy when creating a zone
- **records**: sync PTR records on update and report PTR changes
- **records**: send a change comment with bulk record operations

## [v4.1.0](https://github.com/contentways/poweradmin-go/releases/tag/v4.1.0)

### Features

- add DNSSEC key management and server status clients
- **dnssec**: support active flag on key creation and DNSKEY/DS per key

## [v4.0.0](https://github.com/contentways/poweradmin-go/releases/tag/v4.0.0)

### Features

- **BREAKING**: migrate module path to github.com/contentways/poweradmin-go/v4
- **BREAKING**: use pointer fields for partial updates
- **BREAKING**: support zone transfer when deleting users
- **BREAKING**: support template ID, DNSSEC and ownership on zone create

### Bug Fixes

- accept numeric record IDs from the API
- use the envelope message for API errors
- let releaser-pleaser bump the SDK version constant
- **BREAKING**: send zone updates with the fields the API reads
- return persisted state from zone template updates
- do not retry non-idempotent requests on 5xx or network errors

## [v3.1.0](https://github.com/contentways/poweradmin-go/releases/tag/v3.1.0)

### Features

- add zone metadata endpoints (list, get, set, delete)

### Bug Fixes

- remove dead nil-check in delete-record command

## [v3.0.0](https://github.com/contentways/poweradmin-go/releases/tag/v3.0.0)

### Features

- **BREAKING**: migrate module path to github.com/contentways/poweradmin-go/v3

### Bug Fixes

- correct repository case in releaser-pleaser workflow

## [v3.0.1](https://github.com/contentways/poweradmin-go/releases/tag/v3.0.1)

### Bug Fixes

- correct repository case in releaser-pleaser workflow

## [v2.1.0](https://github.com/Contentways/poweradmin-go/releases/tag/v2.1.0)

### Features

- add DNSSEC support for zones (GetDNSSEC, SetDNSSEC)

## [v1.1.2](https://github.com/Contentways/poweradmin-go/releases/tag/v1.1.2)

### Bug Fixes

- align User-Agent with poweradmin-go project name

## [v1.1.2](https://github.com/Contentways/poweradmin-go/releases/tag/v1.1.2)

### Bug Fixes

- align User-Agent with poweradmin-go project name

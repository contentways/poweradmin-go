#!/bin/sh
# Copyright (c) 2026 contentways
# SPDX-License-Identifier: MIT
set -eu

tool() {
    (
        set -eux
        go run -modfile=../tools/go.mod "$@"
    )
}

# ── ifacemaker: generate interfaces from concrete client structs ──────────────
# Output: zz_*_client_iface.go
# Convention: I<Name>Client (IZoneClient, IRecordClient, ...)

tool github.com/vburenin/ifacemaker \
    -f zone.go       -s ZoneClient       -i IZoneClient       -p poweradmin -o zz_zone_client_iface.go

tool github.com/vburenin/ifacemaker \
    -f record.go     -s RecordClient     -i IRecordClient     -p poweradmin -o zz_record_client_iface.go

tool github.com/vburenin/ifacemaker \
    -f rrset.go      -s RRSetClient      -i IRRSetClient      -p poweradmin -o zz_rrset_client_iface.go

tool github.com/vburenin/ifacemaker \
    -f user.go       -s UserClient       -i IUserClient       -p poweradmin -o zz_user_client_iface.go

tool github.com/vburenin/ifacemaker \
    -f group.go      -s GroupClient      -i IGroupClient      -p poweradmin -o zz_group_client_iface.go

tool github.com/vburenin/ifacemaker \
    -f permission.go -s PermissionClient -i IPermissionClient -p poweradmin -o zz_permission_client_iface.go

tool github.com/vburenin/ifacemaker \
    -f zone_template.go -s ZoneTemplateClient -i IZoneTemplateClient -p poweradmin -o zz_zone_template_client_iface.go

tool github.com/vburenin/ifacemaker \
    -f permission_template.go -s PermissionTemplateClient -i IPermissionTemplateClient -p poweradmin -o zz_permission_template_client_iface.go

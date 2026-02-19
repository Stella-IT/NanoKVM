#!/bin/sh
# Common helpers for network init scripts

# Build udhcpc hostname option from /etc/hostname safely.
# - keep only [A-Za-z0-9.-]
# - trim leading/trailing dots/hyphens
# - drop trailing ".local" (case-insensitive)
# - limit to 253 chars
build_udhcpc_hostname_opt() {
    local _h _s
    _h="$(cat /etc/hostname 2>/dev/null | head -1)"
    _s="$(printf '%s' "$_h" \
        | tr -cd 'A-Za-z0-9.-' \
        | sed 's/^[.-]*//; s/[.-]*$//' \
        | sed -E 's/\\.local$//I' \
        | cut -c1-253)"
    if [ -n "$_s" ]; then
        printf -- "-x hostname:%s" "$_s"
    fi
}


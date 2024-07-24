INSERT INTO firewall (
    "id",
    "name",
    "allow_all",
    "require_auth"
) VALUES (
    "00000000-0000-0000-0000-000000000000",
    'DEFAULT_TST',
    false,
    false
);

INSERT INTO routes (
    "id",
    "name",
    "route",
    "target",
    "firewall_id",
    "include_subroutes"
) VALUES (
    "00000000-0000-0000-0000-000000000001",
    'AUTH',
    'auth',
    'http://localhost:8081',
    "00000000-0000-0000-0000-000000000000",
    true
);

INSERT INTO methods (
    "firewall_id",
    "route_id",
    "method"
) VALUES (
    "00000000-0000-0000-0000-000000000000",
    "00000000-0000-0000-0000-000000000001",
    'GET'
);

-- Localhost: IPv4
INSERT INTO ips (
    "firewall_id",
    "route_id",
    "ip"
) VALUES (
    "00000000-0000-0000-0000-000000000000",
    "00000000-0000-0000-0000-000000000001",
    '127.0.0.1'
);

-- Localhost: IPv6
INSERT INTO ips (
    "firewall_id",
    "route_id",
    "ip"
) VALUES (
    "00000000-0000-0000-0000-000000000000",
    "00000000-0000-0000-0000-000000000001",
    '::1'
);
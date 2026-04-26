SELECT 'CREATE DATABASE ride_hailing_driver'
WHERE NOT EXISTS (
    SELECT 1 FROM pg_database WHERE datname = 'ride_hailing_driver'
)\gexec

SELECT 'CREATE DATABASE ride_hailing_rider'
WHERE NOT EXISTS (
    SELECT 1 FROM pg_database WHERE datname = 'ride_hailing_rider'
)\gexec

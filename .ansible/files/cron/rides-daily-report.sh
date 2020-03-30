#!/usr/bin/bash
start_date=${1:-$(date -d @$(($(date +"%s") - 86400)) +"%Y-%m-%d")}

echo "daily rides report: staring for ${start_date}"

report_dir="/home/proftpd/ntcb/reports/rides-daily/v1/${start_date}"
mkdir -p "${report_dir}"

device_ids_str=$(clickhouse-client --database=tracking --query="SELECT arrayStringConcat(groupArray(toString(device_id)), ' ')
FROM (
      SELECT DISTINCT device_id
      FROM tracking.telemetry
      WHERE nav_timestamp >= toDateTime('${start_date} 00:00:00')
        AND nav_timestamp <= toDateTime('${start_date} 23:59:59')
      ORDER BY device_id);
")

device_ids=(${device_ids_str})

for i in "${device_ids[@]}"; do
  echo "daily rides report: writing report to ${report_dir}/${i}.csv"
  clickhouse-client --database=tracking --query="
  SELECT device_id, nav_timestamp, lat, lon, odometer, fuel_level_liters, ignition_change
FROM (
      WITH
-- get first record with ignition off
          (SELECT timestamp
           FROM tracking.telemetry
           WHERE timestamp >= toDateTime('${start_date} 00:00:00')
             AND timestamp <= toDateTime('${start_date} 23:59:59')
             AND ignition_on = 0
             AND device_id = '${i}'
           ORDER BY timestamp ASC
           LIMIT 1) AS start_date_time,
          (
              SELECT if(
                         -- if the last record for the device_id has ignition_on = 1
                         -- we should lookup when the ignition turns off next day
                                 (SELECT ignition_on
                                  FROM tracking.telemetry
                                  WHERE timestamp >= toDateTime('${start_date} 00:00:00')
                                    AND timestamp <= toDateTime('${start_date} 23:59:59')
                                    AND device_id = '${i}'
                                  ORDER BY timestamp DESC
                                  LIMIT 1) = 1,
                                 (SELECT timestamp
                                  FROM tracking.telemetry
                                  WHERE timestamp <= addHours(toDateTime('${start_date} 23:59:59'), 24)
                                    AND timestamp > toDateTime('${start_date} 23:59:59')
                                    AND device_id = '${i}'
                                    AND ignition_on = 0
                                  ORDER BY timestamp
                                  LIMIT 1),
                                 toDateTime('${start_date} 23:59:59'))
          ) AS end_date_time
      SELECT device_id,
             nav_timestamp,
             lat,
             lon,
             odometer,
             fuel_level_liters,
             runningDifference(ignition_on) AS ignition_change
      FROM tracking.telemetry
      WHERE timestamp >= start_date_time
        AND timestamp <= end_date_time
        AND device_id = '${i}'
        -- we should get the first valid record when ignition turns on
        AND ((odometer > -1 AND fuel_level_liters > -1 AND ignition_on = 1) OR ignition_on = 0)
      ORDER BY timestamp)
WHERE ignition_change <> 0
FORMAT CSVWithNames
" >"${report_dir}/${i}.csv"

done

echo "daily rides report: done"

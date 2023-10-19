package environment

/*
 * # C9500
 * Power                                                    Fan States
 * Supply  Model No              Type  Capacity  Status     0     1
 * ------  --------------------  ----  --------  ---------  -----------
 * PS0     C9K-PWR-650WAC-R      AC    650 W     ok         good  N/A
 * PS1     C9K-PWR-650WAC-R      AC    650 W     fail       N/A   N/A
 *
 * # C9200L
 * SW  PID                 Serial#     Status           Sys Pwr  PoE Pwr  Watts
 * --  ------------------  ----------  ---------------  -------  -------  -----
 * 1A  PWR-C5-600WAC       123          OK              Good     Good     600
 * 1B  Not Present
 */
var templ_power = `# show environment all
Value LOCATION (\w+)
Value MODEL ([\w\-]+)
Value TYPE (\w+)
Value CAPACITY (\d+\s\w+)
Value STATUS ([nN]ot [pP]resent|\w+)

Start
  ^Supply\s+Model -> Power_C9500
  ^SW\s+PID -> Power_C9200L

Power_C9500
  ^${LOCATION}\s+${MODEL}\s+${TYPE}\s+${CAPACITY}\s+${STATUS} -> Record
  ^$ -> End
  
Power_C9200L
  ^${LOCATION}\s+${MODEL}\s+\w+\s+${STATUS} -> Record
  ^${LOCATION}\s+${STATUS} -> Record
  ^$ -> End
`

/*
 * # C9500
 * Sensor List:  Environmental Monitoring
 *  Sensor                  Location        State           Reading
 *  PSOC-MB_0: VOUT         R0              Normal          12116 mV
 *  Temp: Coretemp          R0              Normal          35 Celsius
 *  Temp: OutletDB          R0              Normal          29 Celsius
 *
 * # C9200L
 * Sensor List: Environmental Monitoring
 *  Sensor          Location        State               Reading       Range(min-max)
 *  PS1 Vout        1               GOOD               55000 mV          na
 *  PS1 Hotspot     1               GOOD                  31 Celsius     na
 *  PS1 Fan Status  1               GOOD               43008 rpm         na
 *  PS1 Status word 1               GOOD                   2             na
 *  PS2 Hotspot     1               NOT PRESENT            0 Celsius     na
 *  SYSTEM INLET    1               GREEN                 23 Celsius   0 - 56
 */
var templ_temp = `# show environment all
Value SENSOR ((?:\w+\s?)+)
Value LOCATION (\w+)
Value STATE ((?:\w+\s?)+)
Value VALUE (\d+)

Start
  ^\s*(?:Temp: )?${SENSOR}\s+${LOCATION}\s+${STATE}\s+${VALUE} Celsius -> Record
  ^$ -> End
`

/*
 * # C9500
 * Power                                                    Fan States
 * Supply  Model No              Type  Capacity  Status     0     1
 * ------  --------------------  ----  --------  ---------  -----------
 * PS0     C9K-PWR-650WAC-R      AC    650 W     ok         good  N/A
 * PS1     C9K-PWR-650WAC-R      AC    650 W     fail       N/A   N/A
 *
 * Fan                 Fan States
 * Tray    Status      0     1     2     3
 * ------  ----------  -----------------------
 * FM0     ok          good  good  good  good
 * FM1     ok          good  good  good  good
 *
 * # C9200L
 * Sensor List: Environmental Monitoring
 *  Sensor          Location        State               Reading       Range(min-max)
 *  PS1 Hotspot     1               GOOD                  31 Celsius     na
 *  PS1 Fan Status  1               GOOD               43008 rpm         na
 *  PS2 Hotspot     1               NOT PRESENT            0 Celsius     na
 *  PS2 Fan Status  1               NOT PRESENT            0 rpm         na
 *
 * Switch	 FAN	 Speed	 State	 Airflow direction
 * ---------------------------------------------------
 *   1  	  1    5100 	  OK	 Front to Back
 *   1  	  2	   5145 	  OK	 Front to Back
 *
 */
var templ_fan = `# show environment all
Value Filldown NAME ((?:\w+\s?)+)
Value LOCATION (\w+)
Value STATUS ((?:\w+[\s\/]?)+)
Value STATUS1 ((?:\w+[\s\/]?)+)
Value VALUE (\d+)
Value PWR_MODEL ([\w\-]+)

Start
  ^\s*${NAME}\s+${LOCATION}\s+${STATUS}\s+${VALUE} rpm -> Record
  ^Tray\s+Status -> Fan_C9500
  ^Supply\s+Model -> Fan_C9500_PWR
  ^Switch\s+FAN -> Fan_C9200L

Fan_C9500
  ^${NAME}\s+${STATUS} -> Record
  ^$ -> Start

Fan_C9500_PWR
  ^${LOCATION}\s+${PWR_MODEL}\s+\w+\s+\d+\s\w+\s+\w+\s+${STATUS}\s+${STATUS1} -> Record
  ^$ -> Start

Fan_C9200L
  ^\s*${LOCATION}\s+${NAME}\s+${VALUE}\s+${STATUS} -> Record
  ^$ -> Start
`

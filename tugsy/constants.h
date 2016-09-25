#ifndef CONSTANTS_H
#define CONSTANTS_H

typedef int longitude;
typedef int latitude;

// Sources for our position data. WARNING: These values are used
// in View to refer to columns in sprite sheets.
const int LOCAL_ORIGIN = 0;
const int WEB_ORIGIN = 1;
const int TRAIL = 2;

// Vessel types. WARNING: These values are used in View to
// refer to rows in sprite sheets.
const int RECREATIONAL_VESSEL = 0;
const int PASSENGER_VESSEL = 1;
const int COMMERCIAL_VESSEL = 2;

#endif

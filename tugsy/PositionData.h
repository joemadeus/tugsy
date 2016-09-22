#ifndef POSITIONDATA_H
#define POSITIONDATA_H

#include "constants.h"

using namespace std;

struct vessel_info_t {
    latitude lat;
    longitude lon;
    int heading;
    int vesselType;
    int dataOrigin;
    long timestampMs;
    std::string name;
    std::string designation;
    bool isLatestPosition;
}

class PositionData {

    /**
     * Manage position data, taking input NEMA data from the radio and other
     * sources and, on demand, providing collated lists of vessel positions.
     */

public:
    void updateVessels();
    void expireVessels();

    /**
     * Returns the latest positions of all vessels, sorted
     * first by timestampMs then by vessel designation.
     */
    std::set<vessel_info_t> getLatestPositions(unsigned long const sinceMillis);

    /**
     * Returns the historical positions of all vessels, sorted
     * first by timestampMs then by vessel designation.
     */
    std::set<vessel_info_t> getPastPositions(unsigned long const sinceMillis);
}

#endif

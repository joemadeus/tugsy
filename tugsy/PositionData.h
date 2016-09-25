#ifndef POSITIONDATA_H
#define POSITIONDATA_H

#include <set>
#include <string>
#include "constants.h"

using namespace std;

struct vessel_info_t {
    const latitude lat;
    const longitude lon;
    const unsigned int heading;
    const unsigned int vesselType;
    const unsigned int dataOrigin;
    const unsigned long timestampMs;
    std::string name;
    std::string designation;
    const bool isLatestPosition;
};

const vessel_info_t NULL_VESSEL = { 0, 0, 0, 0, 0, 0, "", "", false };

class PositionData {

    /**
     * Manage position data, taking input NEMA data from the radio and other
     * sources and, on demand, providing collated lists of vessel positions.
     */

public:
    void updateVessels();
    void expireVessels();

    /**
     * Returns the latest positions of all vessels, sorted by lat, then by lon,
     * then by timestampMs.
     */
    std::set<vessel_info_t> getLatestPositions();
};

#endif

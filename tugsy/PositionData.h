#ifndef POSITIONDATA_H
#define POSITIONDATA_H

struct vessel_info_t {
    int lat;
    int lon;
    int vesselType;
    int dataOrigin;
    bool isLatestPosition;
}

class PositionData {
public:
    void updateVessels();
    void expireVessels();
    position_t[] getUpdatedPositions(int sinceMillis);
    position_t[] getExpiredPositions(int sinceMillis);

private:
    unsigned long lastUpdate;
}

#endif

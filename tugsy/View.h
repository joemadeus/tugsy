#ifndef VIEW_H
#define VIEW_H

#include <SDL2/SDL.h>
#include "sdl.h"
#include "PositionData.h"

using namespace std;

const int MARKER_HEIGHT = 8;
const int MARKER_WIDTH = 8;

const int BASE_MARKER_COL = 0;
const int LOCAL_ORIGIN_MARKER_COL = 1;
const int WEB_ORIGIN_MARKER_COL = 2;
const int TRAIL_MARKER_COL = 3;

const int RECREATION_VESSEL_ROW = 0;
const int PASSENGER_VESSEL_ROW = 1;
const int COMMERCIAL_VESSEL_ROW = 2;

const char* BACKGROUND_FILE_NAME = 'view.png';
const char* MARKER_SHEET_FILE_NAME = 'markers.png';

struct marker_dims_t {
    int x;
    int y;
    int w;
    int h;
};

class View {
public:
    View(std::string &name);
    ~View();
    void updateVessel(vessel_info_t vesselData);

private:
    marker_dims_t getMarker(int markerType const, int vesselType const);

    std::string viewName;
    int ulLat;
    int ulLon;
    int lrLat;
    int lrLon;

    SDL_Texture* background;
    SDL_Texture* marker_sheet;
};

#endif

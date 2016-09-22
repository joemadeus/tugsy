#ifndef VIEW_H
#define VIEW_H

#include <array>
#include <iterator>
#include <math.h>
#include <SDL2/SDL.h>

#include "constants.h"
#include "sdl.h"
#include "PositionData.h"

using namespace std;

const std::array<std::string> KNOWN_VIEWS = {
    "pvd_harbor",
    "pvd_to_bristol",
    "pvd_to_gansett"
};

// Both of these should be even multiples of two to make the math work
const int MARKER_HEIGHT = 8;
const int MARKER_WIDTH = 8;

const char* BACKGROUND_FILE_NAME = 'view.png';
const char* MARKER_SHEET_FILE_NAME = 'markers.png';

class View {

    /**
     * Manage how we draw vessels and their trails on an SDL context. Input
     * is a bunch of positon data (see PositionData.h) and output is a bunch
     * of SDL rendering commands.
     */

public:
    View(std::string &name, SdlContext &context);
    ~View();
    void renderPositions(std::set<position_t> latestPositions);

private:
    SDL_Rect getSpriteRect(int dataOrigin const, int vesselType const);
    SDL_Rect getBackgroundRect(latitude lat const, longitude lon const);

    std::string viewName;

    latitude ulLat;
    longitude ulLon;
    latitude lrLat;
    longitude lrLon;

    SdlContext &sdlContext;
    SDL_Texture* background;
    SDL_Texture* marker_sheet;

    std::set<position_t> currentlyRenderedPositions;
};

#endif

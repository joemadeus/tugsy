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

#define NUMBER_OF_KNOWN_VIEWS 3
const std::array<std::string, NUMBER_OF_KNOWN_VIEWS> KNOWN_VIEWS = {
    "pvd_harbor",
    "pvd_to_bristol",
    "pvd_to_gansett"
};

// Both of these should be even multiples of two to make the math work
const int MARKER_HEIGHT = 8;
const int MARKER_WIDTH = 8;

const std::string BACKGROUND_FILE_NAME = "view.png";
const std::string MARKER_SHEET_FILE_NAME = "markers.png";

class View {

    /**
     * Manage how we draw vessels and their trails on an SDL context. Input
     * is a bunch of positon data (see PositionData.h) and output is a bunch
     * of SDL rendering commands.
     */

public:
    View(std::string &name, SdlState &state);
    ~View();
    void rebuild();
    void renderPositions(std::set<vessel_info_t> latestPositions);
    vessel_info_t whatsHere(const int x, const int y);

private:
    void renderPosition(const vessel_info_t &position);
    SDL_Rect* getSpriteRect(const int dataOrigin, const int vesselType);
    SDL_Rect* getBackgroundRect(const latitude lat, const longitude lon);

    std::string viewName;

    latitude ulLat;
    longitude ulLon;
    latitude lrLat;
    longitude lrLon;

    SdlState sdlState;
    SDL_Texture* background;
    SDL_Texture* marker_sheet;

    std::set<vessel_info_t> currentlyRenderedPositions;
};

#endif

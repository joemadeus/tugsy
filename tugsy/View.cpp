#include "View.h"

View::View(std::string &name, SdlState &state) {
    this->viewName = name;
    this->sdlState = state;
    this->background =
        this->sdlState.initTexture(this->viewName, BACKGROUND_FILE_NAME);
    this->marker_sheet =
        this->sdlState.initTexture(this->viewName, MARKER_SHEET_FILE_NAME);

    // TODO: Load the geo coords of the view's four corners from
    // a property file, then use them to assign the local geo vars
}

View::~View() {
    cleanup(this->background, this->marker_sheet);
}

SDL_Rect* View::getSpriteRect(const int dataOrigin, const int vesselType) {
    /**
     * Returns an SDL_Rect that can be used to get a sprite from a
     * vessel/origin sprite sheet.
     */
    SDL_Rect* sdlRect = {};
    sdlRect->x = dataOrigin * MARKER_WIDTH;
    sdlRect->y = dataOrigin * MARKER_HEIGHT;
    sdlRect->w = MARKER_WIDTH;
    sdlRect->h = MARKER_HEIGHT;

    return sdlRect;
}

SDL_Rect* View::getBackgroundRect(const latitude lat, const longitude lon) {
    const int x = (int) round((this->ulLon - lon) / (this->ulLon - this->lrLon));
    const int y = (int) round((this->ulLat - lat) / (this->ulLat - this->lrLat));

    SDL_Rect* sdlRect = {};
    sdlRect->x = x - MARKER_HEIGHT / 2;
    sdlRect->y = y - MARKER_WIDTH / 2;
    sdlRect->h = MARKER_HEIGHT;
    sdlRect->w = MARKER_WIDTH;

    return sdlRect;
}

void View::rebuild() {
    this->sdlState.paint(this->background, NULL, NULL);
    this->currentlyRenderedPositions.clear();
}

void View::renderPosition(const vessel_info_t &position) {
    SDL_Rect* spriteRect =
        this->getSpriteRect(
            position.dataOrigin,
            (position.isLatestPosition) ? position.vesselType : TRAIL);
    SDL_Rect* backgroundRect =
        this->getBackgroundRect(position.lat, position.lon);
    this->sdlState.paint(this->marker_sheet, spriteRect, backgroundRect);
}

void View::renderPositions(std::set<vessel_info_t> latestPositions) {

    // Some technical details: the view maintains a collection of posiition
    // data rendered during the last page flip. On each call to update our
    // view, we run through all the position data, starting from the oldest
    // known point (likely from the last view render) to the most recent
    // (prob'ly brand new data that hasn't even been rendered yet.) If some
    // position data is in the old view but not the new, we re-render the
    // background image to "erase" that point. For the others, we render
    // trail or current-position markers as appropriate. This seems to be a
    // good tradeoff between efficiency and simplicity, in that we're
    // re-rendering all the positions but not the entire view.
    //
    // It's assumed that the position sets are consistently ordered

    std::set<vessel_info_t>::iterator currentIt = this->currentlyRenderedPositions.cbegin();
    std::set<vessel_info_t>::iterator currentEnd = this->currentlyRenderedPositions.cend();
    std::set<vessel_info_t>::iterator latestIt = latestPositions.cbegin();
    std::set<vessel_info_t>::iterator latestEnd = latestPositions.cend();

    while (currentIt != latestIt && currentIt != currentEnd) {
        SDL_Rect* sourceAndDest =
            this->getBackgroundRect(currentIt->lat, currentIt->lon);
        this->sdlState.paint(this->background, sourceAndDest, sourceAndDest);
        currentIt++;
    }

    while (latestIt != latestEnd) {
        this->renderPosition(*latestIt);
        latestIt++;
    }

    this->currentlyRenderedPositions.clear();
    this->currentlyRenderedPositions.swap(latestPositions);
}

vessel_info_t whatsHere(const int x, const int y) {
    return NULL_VESSEL;
}

#include "View.h"

View::View(std::string &viewName, SdlState &context) {
    this->viewName = viewName;
    this->sdlContext = context;
    this->background = context.initTexture(getResource(viewName, BACKGROUND_FILE_NAME));
    this->marker_sheet = context.initTexture(getResource(viewName, MARKER_SHEET_FILE_NAME));

    // TODO: Load the geo coords of the view's four corners from
    // a property file, then use them to assign the local geo vars
}

View::~View() {
    this->sdlContext.cleanup(this->background, this->marker_sheet);
}

SDL_Rect View::getSpriteRect(int dataOrigin const, int vesselType const) {
    /**
     * Returns an SDL_Rect that can be used to get a sprite from a
     * vessel/origin sprite sheet.
     */
    SDL_Rect srcrect;
    srcrect.x = dataOrigin * MARKER_WIDTH;
    srcrect.y = dataOrigin * MARKER_HEIGHT;
    srcrect.w = MARKER_WIDTH;
    srcrect.h = MARKER_HEIGHT;

    return srcrect;
}

SDL_Rect View::getBackgroundRect(latitude lat const, longitude lon const) {
    const int x = (int) round((this->ulLon - lon) / (this->ulLon - this->lrLon));
    const int y = (int) round((this->ulLat - lat) / (this->ulLat - this->lrLat));

    SDL_Rect sdlrect;
    sdlRect.x = x - MARKER_HEIGHT / 2;
    sdlRect.y = y - MARKER_WIDTH / 2;
    sdlRect.h = MARKER_HEIGHT;
    sdlRect.w = MARKER_WIDTH;

    return sdlrect;
}

void View::renderPositions(std::set<position_t> latestPositions) {

    // Some technical details: the view maintains a collection of posiition
    // data rendered during the last page flip. On each call to update our
    // view, we run through all the position data, starting from the oldest
    // known point (likely from the last view render) to the most recent
    // (prob'ly brand new data that hasn't even been rendered yet.) If some
    // position data is in the old view but not the new, we re-render the
    // background image to "erase" that point. For the others, we render
    // trail or current-position markers as appropriate. This seems to be a
    // good tradeoff between efficiency and simplicity, in that we're only
    // re-rendering the position points but not the entire view.
    //
    // it's assumed that the position sets are consistently ordered

    std::set<position_t>::iterator currentIt = this->currentlyRenderedPositions.cbegin();
    std::set<position_t>::iterator currentEnd = this->currentlyRenderedPositions.cend();
    std::set<position_t>::iterator latestIt = latestPositions.cbegin();
    std::set<position_t>::iterator latestEnd = latestPositions.cend();

    while (currentIt != latestIt && currentIt != currentEnd) {
        SDL_Rect sourceAndDest =
            this->getBackgroundRect(currentIt->latitude, currentIt->longitude);
        this->sdlContext.paint(this->background, sourceAndDest, sourceAndDest);
        currentIt++;
    }

    while (latestIt != latestEnd) {
        SDL_Rect spriteRect =
            this->getSpriteRect(
                latestIt->dataOrigin,
                (latestIt->isLatestPosition) ? latestIt->vesselType : TRAIL);
        SDL_Rect backgroundRect =
            this->getBackgroundRect(latestIt->latitude, latestIt->longitude);
        this->sdlContext.paint(this->marker_sheet, spriteRect, backgroundRect);
        latestIt++;
    }

    this->currentlyRenderedPositions.clear();
    this->currentlyRenderedPositions.swap(latestPositions);
}

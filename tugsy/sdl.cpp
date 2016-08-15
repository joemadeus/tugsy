#include <sdl.h>

/**
 * Log an SDL error with some error message to the output stream of our choice
 * @param os The output stream to write the message to
 * @param msg The error message to write, format will be msg error: SDL_GetError()
 */
void logSDLError(std::ostream &os, const std::string &msg){
	os << msg << " error: " << SDL_GetError() << std::endl;
}

void shutdown(sdl_state_t& sdl_state) {
    cleanup(sdl_state.window, sdl_state.renderer);
    IMG_Quit();
    SDL_Quit();
}

int init(sdl_state_t& sdl_state) {
    if (SDL_Init(SDL_INIT_VIDEO) != 0){
    	logSDLError(std::cout, "SDL_Init");
    	return 1;
    }

    sdl_state.window = SDL_CreateWindow(
        "tugsy-dev", SCREEN_X, SCREEN_Y, SCREEN_WIDTH, SCREEN_HEIGHT, SDL_WINDOW_SHOWN);
    if (sdl_state.window == nullptr){
    	logSDLError(std::cout, "CreateWindow");
    	SDL_Quit();
    	return 1;
    }

    sdl_state.renderer = SDL_CreateRenderer(
        sdl_state.window, -1, SDL_RENDERER_ACCELERATED | SDL_RENDERER_PRESENTVSYNC);
    if (sdl_state.renderer == nullptr){
    	logSDLError(std::cout, "CreateRenderer");
    	cleanup(sdl_state.window);
    	SDL_Quit();
    	return 1;
    }

    if ((IMG_Init(IMG_INIT_PNG) & IMG_INIT_PNG) != IMG_INIT_PNG){
    	logSDLError(std::cout, "IMG_Init");
    	SDL_Quit();
    	return 1;
    }

    return 0;
}

std::string getSDLResourcePath() {
    const char PATH_SEP = '/';
    static std::string baseRes;
    if (baseRes.empty()){
		//SDL_GetBasePath will return NULL if something went wrong in retrieving the path
		char *basePath = SDL_GetBasePath();
		if (basePath){
			baseRes = basePath;
			SDL_free(basePath);
		} else {
			std::cerr << "Error getting resource path: " << SDL_GetError() << std::endl;
			return "";
		}
		//We replace the last bin/ with res/ to get the the resource path
		size_t pos = baseRes.rfind("bin");
		baseRes = baseRes.substr(0, pos) + "res" + PATH_SEP;
	}

    return baseRes;
}

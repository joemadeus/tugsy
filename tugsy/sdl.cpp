#include <sdl.h>

SdlState::SdlState() {
	initContext();
	initResourcePath();
}

SdlState::~SdlState() {
	cleanup(this.window, this.renderer);
	IMG_Quit();
	SDL_Quit();
}

void SdlState::initContext() {
	if (SDL_Init(SDL_INIT_VIDEO) != 0){
		logSDLError(std::cout, "SDL_Init");
		// TODO: throw
	}

	this.window = SDL_CreateWindow(
		"tugsy-dev", SCREEN_X, SCREEN_Y, SCREEN_WIDTH, SCREEN_HEIGHT, SDL_WINDOW_SHOWN);
	if (this.window == nullptr){
		logSDLError(std::cout, "CreateWindow");
		SDL_Quit();
		// TODO: throw
	}

	this.renderer = SDL_CreateRenderer(
		this.window, -1, SDL_RENDERER_ACCELERATED | SDL_RENDERER_PRESENTVSYNC);
	if (this.renderer == nullptr){
		logSDLError(std::cout, "CreateRenderer");
		cleanup(this.window);
		SDL_Quit();
		// TODO: throw
	}

	if ((IMG_Init(IMG_INIT_PNG) & IMG_INIT_PNG) != IMG_INIT_PNG){
		logSDLError(std::cout, "IMG_Init");
		SDL_Quit();
		// TODO: throw
	}
}

void SdlState::initResourcePath() {
    const char PATH_SEP = '/';

	// SDL_GetBasePath will return NULL if something went wrong in retrieving the path
	char *basePath = SDL_GetBasePath();
	if (basePath){
		baseRes = basePath;
		SDL_free(basePath);
	} else {
		std::cerr << "Error getting resource path: " << SDL_GetError() << std::endl;
		// TODO: throw
	}

	// We replace the last bin/ with res/ to get the the resource path
	size_t pos = baseRes.rfind("bin");
	baseRes = baseRes.substr(0, pos) + "res" + PATH_SEP;
}

std::string SdlState::getSDLResourcePath() {
	return baseRes;
}

/**
 * Log an SDL error with some error message to the output stream of our choice
 * @param os The output stream to write the message to
 * @param msg The error message to write, format will be msg error: SDL_GetError()
 */
void logSDLError(std::ostream &os, const std::string &msg){
	os << msg << " error: " << SDL_GetError() << std::endl;
}

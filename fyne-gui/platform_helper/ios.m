// +build ios

#import <Foundation/Foundation.h>

void NetworkRequest() {
	[[NSURLSession.sharedSession dataTaskWithURL:[NSURL URLWithString:@"http://captive.apple.com"]] resume];
}

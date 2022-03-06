// +build ios

#import <Foundation/Foundation.h>
#import <AVFoundation/AVFoundation.h>
#include <stdio.h>

void NetworkRequest() {
	[[NSURLSession.sharedSession dataTaskWithURL:[NSURL URLWithString:@"http://captive.apple.com"]] resume];
}

static AVAudioPlayer *player=nil;

void playBackgroundMusic() {
	AVAudioSession *session=[AVAudioSession sharedInstance];
	[session setCategory:AVAudioSessionCategoryPlayback withOptions:AVAudioSessionCategoryOptionMixWithOthers error:nil];
	[session setActive:YES error:nil];
	NSString *musicPath=[[NSBundle mainBundle] pathForResource:@"empty" ofType:@"mp3"];
	NSURL *path=[[NSURL alloc] initFileURLWithPath:musicPath];
	player=[[AVAudioPlayer alloc] initWithContentsOfURL:path error:nil];
	[player prepareToPlay];
	player.numberOfLoops=-1;
	[player play];
}

void stopBackgroundMusic() {
	[player stop];
	player=nil;
	[[AVAudioSession sharedInstance] setActive:NO error:nil];
}
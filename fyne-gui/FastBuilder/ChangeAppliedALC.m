//
//  ChangeAppliedALC.m
//  FastBuilder
//
//  Created by LNSSPsd on 2019/5/10.
//  Copyright © 2019 FastBuilder Dev Group. All rights reserved.
//

#import "ViewController.h"
#import "AppDelegate.h"
#import "ChangeAppliedALC.h"

@interface ChangeAppliedALC ()

@end

@implementation ChangeAppliedALC
- (IBAction)exitapp:(id)sender {
    id delegate=[UIApplication sharedApplication].delegate;
    [delegate exitApp];
}

@end

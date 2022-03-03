function FB_GeneralCmd(fbCmd){
    r=_FB_GeneralCmd(fbCmd)
    if(r instanceof Error){
        throw r
    }
    return r
}

function FB_SendMCCmd(mcCmd){
    r=_FB_SendMCCmd(mcCmd)
    if(r instanceof Error){
        throw r
    }
    return r
}

function FB_SendMCCmdAndGetResult(mcCmd){
    r=_FB_SendMCCmdAndGetResult(mcCmd)
    if(r instanceof Error){
        throw r
    }
    return JSON.parse(r)
}

function FB_RequireUserInput(hint){
    r=_FB_RequireUserInput(hint)
    if(r instanceof Error){
        throw r
    }
    return r
}

function FB_Println(msg){
    r=_FB_Println(msg)
    if(r instanceof Error){
        throw r
    }
    return r
}

function FB_RegPackCallBack(packetType,callBackFn){
    r=_FB_RegPackCallBack(packetType,function (jsonPacket) {
        // console.log(jsonPacket)
        callBackFn(JSON.parse(jsonPacket))
    })
    if (r instanceof Error){
        throw r
    }
    return r
}

// 订阅聊天信息
// 实际上只是对 golang 函数 _FB_RegPackCallBack 的重新利用
function FB_RegChat(callBackFn){
    r=_FB_RegPackCallBack("IDText",function (jsonPacket) {
        chatMsg=JSON.parse(jsonPacket)
        SourceName=chatMsg["SourceName"]
        Message=chatMsg["Message"]
        callBackFn(SourceName,Message)
    })
    if (r instanceof Error){
        throw r
    }
    return r
}

function FB_Query(info){
    r=_FB_Query(info)
    if (r instanceof Error){
        throw r
    }
    return r
}

function FB_SaveFile(fileName,data){
    if (_FB_SaveFile(fileName,data) instanceof Error){
        throw r
    }
}

function FB_ReadFile(fileName){
    r=_FB_ReadFile(fileName)
    if (r instanceof Error){
        throw r
    }
    return r
}

function FB_websocketConnectV1(serverAddress,onMessage) {
    r=_websocketConnectV1(serverAddress,function (newMessage) {
        if(newMessage instanceof Error){
            throw newMessage
        }
        onMessage(newMessage)
    })
    if(r instanceof Error){
        throw r
    }
    return r
}
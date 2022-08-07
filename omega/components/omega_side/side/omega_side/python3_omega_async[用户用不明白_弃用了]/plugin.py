from abc import abstractmethod
import asyncio
from typing import Awaitable
from .core import MainFrame
from .core import get_mainframe


class BasicPlugin(object):
    def __init__(self) -> None:
        pass
    
    @abstractmethod
    async def __call__(self) -> Awaitable:
        pass
        
    @property
    def frame(self) -> MainFrame:
        return get_mainframe()
    
async def basic_plugin(frame:MainFrame) -> Awaitable:
    pass 
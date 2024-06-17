import logging
import os
import PyMCTranslate
import itertools
import amulet_nbt as nbt
from typing import Optional, Any, List
from PyMCTranslate.py3.api import Block

print_extra_needed = True

translations = PyMCTranslate.new_translation_manager()

def get_blockstates(version, namespace_str, base_name,force_blockstate):
    block_specification = version.block.get_specification(
        namespace_str, base_name, force_blockstate
    )
    properties = block_specification.get("properties", {})
    if len(properties) > 0:
        keys, values = zip(*properties.items())
    else:
        keys, values = (), ()
    values = tuple([nbt.from_snbt(val) for val in prop] for prop in values)

    for spec_ in itertools.product(*values):
        spec = dict(zip(keys, spec_))
        yield Block(namespace=namespace_str, base_name=base_name, properties=spec)

out = translations.get_version("bedrock", tuple(int(i) for i in "1.20.10".split(".")))

snbt_fp=open("snbt_convert.txt","w")

no_convert_blocks={}
# for namespace_str in out.block.namespaces(True):
#     for base_name in out.block.base_names(namespace_str, True):
#         for input_blockstate in get_blockstates(out, namespace_str, base_name,True):
#             block_str=f"{input_blockstate}"
#             no_convert_blocks[block_str]=True


known_blocks={}
for platform_name in translations.platforms():
    for version_number in translations.version_numbers(platform_name):
        # if f"{platform_name}_{version_number}"!="bedrock_(1, 1, 0)":
        #     continue
        print(f"{platform_name}_{version_number}")
        version = translations.get_version(platform_name, version_number)
        has_blockdata=True
        for namespace_str in version.block.namespaces(False):
            for base_name in version.block.base_names(namespace_str, False):
                for input_blockstate in get_blockstates(version, namespace_str, base_name,False):
                    has_blockdata= "block_data" in input_blockstate.properties
                    break
                break
            break
        
        for namespace_str in version.block.namespaces(True):
            for base_name in version.block.base_names(namespace_str, True):
                for input_blockstate in get_blockstates(version, namespace_str, base_name,True):
                    block_str=f"{input_blockstate}"
                    if block_str not in no_convert_blocks and block_str not in known_blocks:
                        known_blocks[block_str]=input_blockstate
                        universal_output, extra_output, extra_needed = version.block.to_universal(input_blockstate, force_blockstate=True)
                        if extra_needed:
                            print(universal_output)
                        outblock,extra_output, extra_needed = out.block.from_universal(universal_output, force_blockstate=True)
                        out_str=f"{outblock}"
                        snbt_fp.write(f"in : {block_str}\nuni: {universal_output}\nout: {out_str}\n")
        
        if has_blockdata:
            for namespace_str in version.block.namespaces(False):
                for base_name in version.block.base_names(namespace_str, False):
                    for input_blockstate in get_blockstates(version, namespace_str, base_name,False):
                        block_str=f"{input_blockstate}"
                        if block_str not in no_convert_blocks and block_str not in known_blocks:
                            known_blocks[block_str]=input_blockstate
                            universal_output, extra_output, extra_needed = version.block.to_universal(input_blockstate, force_blockstate=True)
                            if extra_needed:
                                print(universal_output)
                            outblock,extra_output, extra_needed = out.block.from_universal(universal_output, force_blockstate=True)
                            out_str=f"{outblock}"
                            snbt_fp.write(f"in : {block_str}\nuni: {universal_output}\nout: {out_str}\n")   


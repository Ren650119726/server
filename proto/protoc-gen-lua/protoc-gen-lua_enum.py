#!/usr/bin/env python
# Only Enum  by Boyee 2017.8.1
# as listed at <url: http://www.opensource.org/licenses/bsd-license.php >.

import sys
import os.path as path
from cStringIO import StringIO

import plugin_pb2
import google.protobuf.descriptor_pb2 as descriptor_pb2
 
_files = {} 

FDP = plugin_pb2.descriptor_pb2.FieldDescriptorProto

if sys.platform == "win32":
    import msvcrt
    import os
    msvcrt.setmode(sys.stdin.fileno(), os.O_BINARY)
    msvcrt.setmode(sys.stdout.fileno(), os.O_BINARY)
	 
def printerr(*args):
    sys.stderr.write(" ".join(args))
    sys.stderr.write("\n")
    sys.stderr.flush()

class TreeNode(object):
    def __init__(self, name, parent=None, filename=None, package=None):
        super(TreeNode, self).__init__()
        self.child = []
        self.parent = parent
        self.filename = filename
        self.package = package
        if parent:
            self.parent.add_child(self)
        self.name = name

    def add_child(self, child):
        self.child.append(child)

    def find_child(self, child_names):
        if child_names:
            for i in self.child:
                if i.name == child_names[0]:
                    return i.find_child(child_names[1:])
            raise StandardError
        else:
            return self

    def get_child(self, child_name):
        for i in self.child:
            if i.name == child_name:
                return i
        return None

    def get_path(self, end=None):
        pos = self
        out = []
        while pos and pos != end:
            out.append(pos.name)
            pos = pos.parent
        out.reverse()
        return '.'.join(out)

    def get_global_name(self):
        return self.get_path()

    def get_local_name(self):
        pos = self
        while pos.parent:
            pos = pos.parent
            if self.package and pos.name == self.package[-1]:
                break
        return self.get_path(pos)

    def __str__(self):
        return self.to_string(0)

    def __repr__(self):
        return str(self)

    def to_string(self, indent=0):
        return ' ' * indent + '<TreeNode ' + self.name + '(\n' + \
                ','.join([i.to_string(indent + 4) for i in self.child]) + \
                ' ' * indent + ')>\n'

class Env(object):
    filename = None
    package = None
    extend = None
    descriptor = None
    message = None
    context = None
    register = None
    def __init__(self):
        self.message_tree = TreeNode('')
        self.scope = self.message_tree

    def get_global_name(self):
        return self.scope.get_global_name()

    def get_local_name(self):
        return self.scope.get_local_name()

    def get_ref_name(self, type_name):
        try:
            node = self.lookup_name(type_name)
        except:
            # if the child doesn't be founded, it must be in this file
            return type_name[len('.'.join(self.package)) + 1:]
        if node.filename != self.filename:
            return node.filename + '_pb.' + node.get_local_name()
        return node.get_local_name()

    def lookup_name(self, name):
        names = name.split('.')
        if names[0] == '':
            return self.message_tree.find_child(names[1:])
        else:
            return self.scope.parent.find_child(names)

    def enter_package(self, package):
        if not package:
            return self.message_tree
        names = package.split('.')
        pos = self.message_tree
        for i, name in enumerate(names):
            new_pos = pos.get_child(name)
            if new_pos:
                pos = new_pos
            else:
                return self._build_nodes(pos, names[i:])
        return pos

    def enter_file(self, filename, package):
        self.filename = filename
        self.package = package.split('.')
        self._init_field()
        self.scope = self.enter_package(package)

    def exit_file(self):
        self._init_field()
        self.filename = None
        self.package = []
        self.scope = self.scope.parent

    def enter(self, message_name):
        self.scope = TreeNode(message_name, self.scope, self.filename,
                              self.package)

    def exit(self):
        self.scope = self.scope.parent

    def _init_field(self):
        self.descriptor = []
        self.context = []
        self.message = []
        self.register = []

    def _build_nodes(self, node, names):
        parent = node
        for i in names:
            parent = TreeNode(i, parent, self.filename, self.package)
        return parent

class Writer(object):
    def __init__(self, prefix=None):
        self.io = StringIO()
        self.__indent = ''
        self.__prefix = prefix

    def getvalue(self):
        return self.io.getvalue()

    def __enter__(self):
        self.__indent += '    '
        return self

    def __exit__(self, type, value, trackback):
        self.__indent = self.__indent[:-4]

    def __call__(self, data):
        self.io.write(self.__indent)
        if self.__prefix:
            self.io.write(self.__prefix)
        self.io.write(data)
		 
def code_gen_file(proto_file, env, is_gen):
    filename = path.splitext(proto_file.name)[0]
    env.enter_file(filename, proto_file.package)
	  

    env.message.append('local %s = {}\n' % (env.filename))
     
    for enum_desc in proto_file.enum_type:        
        for enum_value in enum_desc.value:
            env.message.append('%s.%s = %d\n' % (env.filename,enum_value.name,enum_value.number))
			             
    env.message.append('return %s' % (env.filename))

    if is_gen:
        lua = Writer()         
        map(lua, env.message) 
        _files[env.filename + '.lua'] = lua.getvalue()
    env.exit_file()

def main():
    plugin_require_bin = sys.stdin.read()
    code_gen_req = plugin_pb2.CodeGeneratorRequest()
    code_gen_req.ParseFromString(plugin_require_bin)

    env = Env()
    for proto_file in code_gen_req.proto_file:
        code_gen_file(proto_file, env, proto_file.name in code_gen_req.file_to_generate)

    code_generated = plugin_pb2.CodeGeneratorResponse()
    for k in  _files:
        file_desc = code_generated.file.add()
        file_desc.name = k
        file_desc.content = _files[k]

    sys.stdout.write(code_generated.SerializeToString())

if __name__ == "__main__":
    main()


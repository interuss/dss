import copy
parents = {
    'd1': {'dss'},
    'd2': {'dss'},
    'd3': {'dss'},
    'd4': {'dss'},
    'd5': {'d1', 'd3'},
    'd6': {'d1'},
    'd7': {'d3'},
    'd8': {'d2', 'd3'},
    'd9': {'d4'},
    'd10': {'d4'},
    'd11': {'d5'},
}

def get_path(dest, base, path, all_paths):
    if dest not in path:
        path.append(dest)
    if dest in parents:
        for parent in parents[dest]:
            curr_path = copy.deepcopy(path)
            curr_path.append(parent)
            if parent == base:
                all_paths.append(curr_path)
            else:
                get_path(parent, base, curr_path, all_paths)

def get_all_paths():
    base = 'dss'
    dest = 'd8'
    all_paths = []
    get_path(dest, base, [], all_paths)
    print('all_paths: ', all_paths)

if __name__ == '__main__':
    get_all_paths()
"""Regenerate the planner's searchable icon index from a local Status.csv.

Source: https://github.com/thewakingsands/ffxiv-datamining-cn/blob/master/Status.csv
Run: python3 scripts/update_board_buffs.py /path/to/Status.csv
Only metadata is generated; desktop and WASM reuse the existing ability artwork.
"""

import csv
import sys
from pathlib import Path


def main(status_file):
    root = Path(__file__).resolve().parents[1]
    # Unlike square action icons, status icons live in the 2xxxxx group.
    icons = {
        int(path.stem.split("-")[1]): []
        for path in (root / "asset/abilities").glob("2?????-??????.png")
    }
    with open(status_file, encoding="utf-8-sig", newline="") as src:
        rows = csv.reader(src)
        next(rows)  # column numbers
        columns = next(rows)
        name, icon = columns.index("Name"), columns.index("Icon")
        next(rows)  # field types
        for row in rows:
            names = icons.get(int(row[icon]))
            label = " ".join(row[name].split())
            if names is not None and label and label not in names:
                names.append(label)
    with (root / "internal/planner/buff_icons.tsv").open("w", encoding="utf-8") as out:
        for icon, names in sorted(icons.items()):
            print(icon, *names, sep="\t", file=out)


if __name__ == "__main__":
    main(sys.argv[1])

# Connect-and-errors guide

The guide under Help (`/help/guide`) is generated from Markdown.

- `guide.md`: the guide itself.
- `additions.md`: extra tool sections. Each one starts with `<!-- chapter: 三 -->` naming the chapter it is appended to, and is marked "新增" on the page.
- `build.py`: turns the two files into the data the page renders.
- `template.html`: a standalone page with the same content, for sharing as a single file.

## Regenerate after editing the Markdown

Needs Python 3 and `pip install markdown`.

```bash
cd tools/yc-guide
python3 build.py guide.md ../../frontend/src/data/yc-guide.json additions.md 1,12,16,20,21,24,29,34,35
```

The last argument lists the numbers of the tools in `guide.md` that carry the "已更新" mark; leave it out for none.

For the standalone page, give an `.html` output path instead:

```bash
python3 build.py guide.md guide.html additions.md 1,12,16,20,21,24,29,34,35
```

The page itself lives in `frontend/src/components/yecai/guide/guideApp.ts` and `frontend/src/styles/yc-guide.css`.

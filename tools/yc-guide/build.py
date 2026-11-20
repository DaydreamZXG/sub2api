#!/usr/bin/env python3
"""Builds one self-contained page from the guide's Markdown: python3 build.py guide.md out.html"""
import html, json, re, sys
import markdown

src, out = sys.argv[1], sys.argv[2]
lines = open(src, encoding='utf-8').read().split('\n')
# Optional third argument: sections to add, each preceded by `<!-- chapter: 三 -->`. They are shown as new.
ADDED = {}
if len(sys.argv) > 3:
    parts = re.split(r'<!-- chapter: (.) -->\n', open(sys.argv[3], encoding='utf-8').read())
    for numeral, text in zip(parts[1::2], parts[2::2]):
        ADDED.setdefault(numeral, []).append(text.strip())

# Optional fourth argument: comma-separated numbers of the document's own tools whose text was corrected.
UPDATED = {int(n) for n in sys.argv[4].split(',')} if len(sys.argv) > 4 else set()
# Added sections are inserted at the end of their chapter and numbered after the document's own tools.
BASE_MAX = max(int(m) for m in re.findall(r'^## (\d+)\\?\.', '\n'.join(lines), re.M))
next_num = BASE_MAX
for numeral, sections in ADDED.items():
    start = next(i for i, l in enumerate(lines) if re.match(rf'^# [^\n]*{numeral}、', l))
    fenced, end = False, len(lines)
    for i in range(start + 1, len(lines)):
        if lines[i].startswith('```'):
            fenced = not fenced
        if not fenced and lines[i].startswith('# '):
            end = i
            break
    block = []
    for text in sections:
        next_num += 1
        block += re.sub(r'^## ', f'## {next_num}\\. ', text, count=1).split('\n') + ['', '---', '']
    lines[end:end] = block

STATUS = [('✅ 直接支持', 'direct'), ('🔄 协议转换', 'convert'), ('⚠️ 有限支持', 'limited'), ('❌ 不建议 / 不能直连', 'no')]
unesc = lambda s: re.sub(r'\\([.\-+|_*#!()\[\]{}>~])', r'\1', s)
plain = lambda h: html.unescape(re.sub(r'<[^>]+>', '', h)).strip()
no_emoji = lambda s: re.sub(r'^[^\w\u4e00-\u9fff]+', '', unesc(s)).strip()


def render(text):
    h = markdown.markdown(text, extensions=['tables', 'fenced_code', 'sane_lists'])
    h = re.sub(r'<a href="(https?://[^"]+)"', r'<a target="_blank" rel="noopener" href="\1"', h)
    h = re.sub(r'<blockquote>\s*<p>(⚠️|❗)', r'<blockquote class="warn"><p>\1', h)
    h = h.replace('<table>', '<div class="tbl"><table>').replace('</table>', '</table></div>')
    return h


def table_rows(h):
    """The first table of a rendered block as (head cells, body rows, html before, html after)."""
    m = re.search(r'<div class="tbl"><table>(.*?)</table></div>', h, re.S)
    if not m:
        return None
    rows = [re.findall(r'<t[hd][^>]*>(.*?)</t[hd]>', r, re.S) for r in re.findall(r'<tr>(.*?)</tr>', m.group(1), re.S)]
    return rows[0], rows[1:], h[:m.start()], h[m.end():]


def error_list(h):
    """Turns a three-column 'what you see / why / what to do' table into a list of cards."""
    t = table_rows(h)
    if not t or len(t[0]) != 3:
        return h, []
    head, rows, before, after = t
    errors = [{'s': r[0], 'c': r[1], 'f': r[2]} for r in rows if len(r) == 3]
    cards = ''.join(
        f'<div class="err"><div class="err-s">{e["s"]}</div><dl><dt>{head[1]}</dt><dd>{e["c"]}</dd><dt>{head[2]}</dt><dd>{e["f"]}</dd></dl></div>'
        for e in errors)
    return f'{before}<div class="errs">{cards}</div>{after}', errors


def inline_errors(h):
    """Rows of every 'what you see' table in a block that keeps its error tables inline, for the error search."""
    found = []
    while True:
        t = table_rows(h)
        if not t:
            return found
        head, rows, _, h = t
        if len(head) == 3 and '你看到什么' in head[0]:
            found += [{'s': r[0], 'c': r[1], 'f': r[2]} for r in rows if len(r) == 3]


def kind_of(t):
    if '报错' in t:
        return 'errors'
    if t.startswith('注意事项') or '三坑' in t:
        return 'notes'
    return 'steps'


# Split into blocks at headings that are outside fenced code.
blocks, fenced = [], False
for line in lines:
    if line.startswith('```'):
        fenced = not fenced
    m = None if fenced else re.match(r'^(#{1,2}) (.+)$', line)
    if m:
        blocks.append({'level': len(m.group(1)), 'title': m.group(2).strip(), 'body': []})
    elif blocks:
        blocks[-1]['body'].append(line)
for b in blocks:
    b['text'] = re.sub(r'\n-{3,}\s*$', '', '\n'.join(b['body']).strip()).strip()

title = no_emoji(blocks[0]['title'])
intro = blocks[0]['text']
meta = [re.sub(r'^>\s*', '', l) for l in intro.split('\n') if l.startswith('>') or re.match(r'^[🎯💡]', l)]
intro_rest = '\n'.join(l for l in intro.split('\n') if re.sub(r'^>\s*', '', l) not in meta)

start, tools, pages, general = [], [], [], None
chapter = None
for b in blocks[1:]:
    if b['level'] == 1:
        name = re.sub(r'^[一二三四五六七八九十]+、', '', no_emoji(b['title']))
        chapter = {'id': f'c{len(pages) + 1}', 'title': name, 'text': b['text'], 'subs': []}
        chapter['is_tools'] = False
        pages.append(chapter)
        continue
    if chapter is None:
        start.append({'id': f's{len(start) + 1}', 'title': no_emoji(b['title']), 'html': render(b['text'])})
        continue
    m = re.match(r'^(\d+)\\?\.\s*(.+)$', b['title'])
    if not m:
        chapter['text'] += f"\n\n## {b['title']}\n\n{b['text']}"
        continue
    chapter['is_tools'] = True
    num, rest = int(m.group(1)), m.group(2).strip()
    status = ''
    for mark, key in STATUS:
        if rest.endswith(mark):
            status, rest = key, rest[:-len(mark)].strip()
    rest, note = unesc(rest), ''
    m2 = re.match(r'^(.*?)（([^（）]*)）$', rest)
    if m2:
        rest, note = m2.group(1).strip(), m2.group(2).strip('！!')

    # Segments: a line that is only **bold** starts one.
    segs, cur, fenced, refs = [], {'title': '', 'lines': []}, False, []
    for line in b['text'].split('\n'):
        if line.startswith('```'):
            fenced = not fenced
        if not fenced and re.match(r'^\*\*[^*]+\*\*\s*$', line):
            segs.append(cur)
            cur = {'title': unesc(line.strip().strip('*')), 'lines': []}
        elif not fenced and line.startswith('**官方参考**'):
            refs.append(line)
        elif not fenced and cur['title'] and 'kind' not in cur and re.match(r'^### ', line) and kind_of(cur['title']) != 'steps':
            segs.append(cur)
            cur = {'title': '', 'lines': [line], 'kind': 'more'}
        else:
            cur['lines'].append(line)
    segs.append(cur)
    tools.append({'num': num, 'name': rest, 'note': note, 'status': status, 'cat': chapter['id'], 'segs': segs, 'refs': refs, 'new': num > BASE_MAX, 'updated': num in UPDATED})

LABEL = {'steps': '怎么配置', 'notes': '注意事项', 'errors': '常见报错', 'more': '更多说明'}
out_tools = []
for t in tools:
    first = t['segs'][0]
    paras = '\n'.join(first['lines']).strip().split('\n\n')
    lead, pre = paras[0], '\n\n'.join(paras[1:]).strip()
    tabs, errors = {}, []
    if pre:
        tabs['steps'] = render(pre)
        errors += inline_errors(tabs['steps'])
    for s in t['segs'][1:]:
        kind = s.get('kind') or kind_of(s['title'])
        body = render('\n'.join(s['lines']).strip())
        if kind == 'errors':
            body, errs = error_list(body)
            errors += errs
        else:
            errors += inline_errors(body)
        plain_title = s['title'] in ('配置步骤', '注意事项', '常见报错速查', '')
        head = '' if plain_title else f'<h3>{html.escape(s["title"])}</h3>'
        tabs[kind] = tabs.get(kind, '') + head + body
    out_tools.append({
        'num': t['num'], 'name': t['name'], 'note': t['note'], 'status': t['status'], 'cat': t['cat'], 'new': t['new'], 'updated': t['updated'],
        'lead': plain(render(lead)), 'leadHtml': render(lead),
        'tabs': [{'kind': k, 'label': LABEL[k], 'html': tabs[k]} for k in ('steps', 'notes', 'errors', 'more') if k in tabs],
        'errors': [{'s': e['s'], 'c': e['c'], 'f': e['f'], 't': plain(e['s'] + ' ' + e['c'] + ' ' + e['f']).lower()} for e in errors],
        'refs': render('\n\n'.join(t['refs'])) if t['refs'] else '',
    })

out_cats, out_pages = [], []
for c in pages:
    if c['is_tools']:
        out_cats.append({'id': c['id'], 'title': c['title'], 'html': render(c['text']) if c['text'] else ''})
    elif '报错' in c['title']:
        body, errs = error_list(render(c['text']))
        t = table_rows(render(c['text']))
        general = {'title': c['title'], 'rest': t[3] if t else '',
                   'errors': [{'s': e['s'], 'c': e['c'], 'f': e['f'], 't': plain(e['s'] + ' ' + e['c'] + ' ' + e['f']).lower()} for e in errs]}
    else:
        out_pages.append({'id': c['id'], 'title': c['title'], 'html': render(c['text'])})

data = {'title': title, 'meta': [unesc(m) for m in meta], 'intro': render(intro_rest), 'start': start,
        'cats': out_cats, 'tools': out_tools, 'general': general, 'pages': out_pages}
if out.endswith('.json'):
    # Data only: what the site's Help page renders (frontend/src/data/yc-guide.json).
    open(out, 'w', encoding='utf-8').write(json.dumps(data, ensure_ascii=False, separators=(',', ':')) + '\n')
    print(f"{len(out_tools)} tools, {len(out_cats)} categories, {len(out_pages)} pages -> {out}")
    sys.exit(0)
page = open(__file__.replace('build.py', 'template.html'), encoding='utf-8').read()
page = page.replace('__TITLE__', html.escape(title)).replace('__DATA__', json.dumps(data, ensure_ascii=False).replace('</', '<\\/'))
open(out, 'w', encoding='utf-8').write(page)
print(f"{len(out_tools)} tools, {sum(len(t['errors']) for t in out_tools)} tool errors, {len(general['errors'])} general errors, "
      f"{len(out_cats)} categories, {len(out_pages)} pages, {len(start)} start sections, {len(page) // 1024} KB")
for t in out_tools:
    print(t['num'], t['name'], '|', t['note'], '|', t['status'] or '-', '|', [x['kind'] for x in t['tabs']], len(t['errors']))

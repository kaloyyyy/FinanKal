"""Functions-only graph view, laid out by service. Run from the repo root with graphify's Python.
Reads graph.json and writes graph-by-service.html; re-run after a graph rebuild."""
import json
import re
from pathlib import Path

import networkx as nx
from graphify.export import to_html

SERVICES = [("Svelte", "finankal-ui/"), ("Spring Boot", "api-spring/"), ("Go", "engine-go/")]
# Columns, left to right in call order: (service index, heading).
LAYERS = [
    (0, "Pages & components"), (0, "API client"),
    (1, "Controllers"), (1, "Services"), (1, "Mappers, DTOs, config"),
    (2, "gRPC handlers"), (2, "Domain logic"), (2, "Repositories"), (2, "Cache & DB"),
]
GENERATED = (".pb.go",)  # protobuf output: 343 accessors that would bury the hand-written Go
COL_W, FN_X, FN_H, HEADER_H, GAP, SERVICE_GAP, SWEEPS = 360, 150, 58, 44, 40, 160, 12
NOISE_PREFIXES = ("finankal-ui/src/", "finankal-ui/", "api-spring/src/main/java/com/finankal/api/", "api-spring/src/main/java/com/finankal/", "api-spring/", "engine-go/internal/", "engine-go/")


def short_path(path):
    # Same-named files (+page.svelte, service.go) need their folder to be told apart.
    return next((path[len(p):] for p in NOISE_PREFIXES if path.startswith(p)), path)


def service_of(path):
    return next(i for i, (_, prefix) in enumerate(SERVICES) if path.startswith(prefix))


def layer_of(path):
    if path.startswith("finankal-ui/"):
        return 1 if "/lib/server/api/" in path else 0
    if path.startswith("api-spring/"):
        return 2 if "/controller/" in path else 3 if "/service/" in path else 4
    if "/cmd/" in path:
        return 5
    if "repository" in path.rsplit("/", 1)[-1]:
        return 7
    return 8 if "/internal/cache/" in path or "/internal/db/" in path else 6


assert [layer_of(p) for p in ("finankal-ui/src/routes/+page.svelte", "finankal-ui/src/lib/server/api/users.ts", "api-spring/src/x/controller/A.java", "api-spring/src/x/service/A.java", "api-spring/src/x/mapper/A.java", "engine-go/cmd/server/server.go", "engine-go/internal/ledger/service.go", "engine-go/internal/ledger/repository.go", "engine-go/internal/cache/keys.go")] == list(range(9))


def stack(blocks):
    """Lay one column's ordered file blocks top to bottom. Returns {id: y} and [(path, header y)]."""
    ys, heads, y = {}, [], 0
    for path, ids in blocks:
        heads.append((path, y))
        for i, node_id in enumerate(ids):
            ys[node_id] = y + HEADER_H + i * FN_H
        y += HEADER_H + len(ids) * FN_H + GAP
    return ys, heads


def layered_layout(columns, G, nodes):
    """Barycenter sweeps: order functions and file blocks by the mean height of what they connect to.
    Files stay together. Anything with no links outside its own file sinks to the bottom of its column."""
    order = [sorted((path, sorted(ids)) for path, ids in col.items()) for col in columns]
    y = {}
    for blocks in order:
        y.update(stack(blocks)[0])

    def pull(node_id, path):
        others = [y[m] for m in G.neighbors(node_id) if nodes[m]["source_file"] != path]
        return sum(others) / len(others) if others else None

    for sweep in range(SWEEPS):
        indices = range(len(order)) if sweep % 2 == 0 else reversed(range(len(order)))
        for c in indices:
            ranked = []
            for path, ids in order[c]:
                pulls = {i: pull(i, path) for i in ids}
                known = [p for p in pulls.values() if p is not None]
                # Unlinked functions go after the linked ones; the sort is stable, so ties keep their order.
                ids = sorted(ids, key=lambda i: (pulls[i] is None, pulls[i] or 0))
                ranked.append(((not known, sum(known) / len(known) if known else 0), path, ids))
            order[c] = [(path, ids) for _key, path, ids in sorted(ranked, key=lambda r: r[0])]
            y.update(stack(order[c])[0])
    return order, y


def patch(html, old, new):
    assert html.count(old) == 1, f"graph.html template changed, cannot find: {old!r}"
    return html.replace(old, new)


def route_key(path):
    """/api/users/${id}/x and /api/users/{userId}/x compare equal; a trailing slash or query does not matter."""
    path = re.sub(r"\$\{[^}]*\}|\{[^}]*\}", "{}", path.split("?")[0])
    return path.rstrip("/") or "/"


assert route_key("/api/users/${userId}/transactions/recent?limit=${n}") == route_key("/api/users/{id}/transactions/recent") == "/api/users/{}/transactions/recent"


def line_of(text, offset):
    return text.count("\n", 0, offset) + 1


def enclosing(by_file, path, line):
    """Function whose start line is the closest one at or above `line`."""
    starts = [(start, node_id) for start, node_id in by_file.get(path, []) if start <= line]
    return max(starts)[1] if starts else None


def source_files(prefix, suffixes):
    return sorted(p for p in Path(prefix).rglob("*") if p.suffix in suffixes and p.is_file() and "node_modules" not in p.parts and ".svelte-kit" not in p.parts)


def link_services(nodes):
    """Static text match of the two network hops: fetch URL -> Spring route, gRPC stub call -> Go server method."""
    by_file, by_label = {}, {}
    for node_id, n in nodes.items():
        by_file.setdefault(n["source_file"], []).append((int(n["source_location"].lstrip("L")), node_id))
        by_label.setdefault((n["source_file"], n["label"]), []).append(node_id)

    routes = {}
    for path in source_files("api-spring/src/main/java/com/finankal/api/controller", {".java"}):
        text, rel = path.read_text(encoding="utf-8"), path.as_posix()
        base = re.search(r'@RequestMapping\("([^"]*)"\)', text).group(1)
        found = re.findall(r'@(Get|Post|Put|Delete|Patch)Mapping(?:\("([^"]*)"\))?\s+public [^(]*?(\w+)\(', text)
        assert len(found) == len(re.findall(r"@(?:Get|Post|Put|Delete|Patch)Mapping", text)), f"unparsed route in {rel}"
        for verb, sub, name in found:
            ids = by_label[(rel, f".{name}()")]
            assert len(ids) == 1, f"overloaded controller method {rel} {name}"
            routes[(verb.upper(), route_key(base + sub))] = ids[0]

    edges, unmatched_urls, used_routes = [], [], set()
    for path in source_files("finankal-ui/src", {".ts", ".svelte", ".js"}):
        text, rel = path.read_text(encoding="utf-8"), path.as_posix()
        for m in re.finditer(r"`(?:\$\{\w+\})?(/api/[^`]*)`", text):
            verb = re.search(r'method:\s*"(\w+)"', text[m.end():m.end() + 120])
            key = (verb.group(1).upper() if verb else "GET", route_key(m.group(1)))
            caller = enclosing(by_file, rel, line_of(text, m.start()))
            if key in routes and caller:
                edges.append((caller, routes[key], f"http {key[0]} {key[1]}"))
                used_routes.add(key)
            else:
                unmatched_urls.append((rel, *key))

    go_file = "engine-go/cmd/server/server.go"
    unimplemented, used_rpcs = [], set()
    for path in source_files("api-spring/src/main/java/com/finankal/api/service", {".java"}):
        text, rel = path.read_text(encoding="utf-8"), path.as_posix()
        for m in re.finditer(r"financeEngineStub\s*\.\s*(\w+)\(", text):
            rpc = m.group(1)[0].upper() + m.group(1)[1:]
            caller, target = enclosing(by_file, rel, line_of(text, m.start())), by_label.get((go_file, f".{rpc}()"))
            if target and caller:
                edges.append((caller, target[0], f"grpc {rpc}"))
                used_rpcs.add(rpc)
            else:
                unimplemented.append((rel, rpc))

    unused_routes = sorted(k for k in routes if k not in used_routes)
    all_rpcs = {n["label"].strip(".()") for n in nodes.values() if n["source_file"] == go_file and n["label"].startswith(".")}
    return edges, unmatched_urls, unused_routes, unimplemented, sorted(all_rpcs - used_rpcs)


graph = json.loads(Path("graphify-out/graph.json").read_text(encoding="utf-8"))
nodes = {
    n["id"]: n for n in graph["nodes"]
    if n["label"].endswith("()") and not n["source_file"].endswith(GENERATED)
}

G = nx.Graph()
for node_id, n in nodes.items():
    G.add_node(node_id, **{k: v for k, v in n.items() if k not in ("id", "community", "community_name")})
for e in graph["links"]:
    if e["source"] in nodes and e["target"] in nodes:
        # _src/_tgt keep the arrow caller -> callee, since an undirected Graph may swap the pair.
        G.add_edge(e["source"], e["target"], _src=e["source"], _tgt=e["target"], **{k: v for k, v in e.items() if k not in ("source", "target", "_src", "_tgt")})

network_edges, unmatched_urls, unused_routes, unimplemented, uncalled_rpcs = link_services(nodes)
for a, b, relation in network_edges:
    G.add_edge(a, b, _src=a, _tgt=b, relation=relation, confidence="EXTRACTED", confidence_score=1.0, weight=1.0)

# Service index doubles as the community, so colours and the legend follow the services.
communities = {i: [] for i in range(len(SERVICES))}
columns = [{} for _ in LAYERS]
for node_id, n in nodes.items():
    i = service_of(n["source_file"])  # StopIteration here means a function outside the three services
    communities[i].append(node_id)
    columns[layer_of(n["source_file"])].setdefault(n["source_file"], []).append(node_id)

order, ys = layered_layout(columns, G, nodes)
positions, headers, subtitles, titles, left, seen = {}, [], [], [], 0, None
for c, (service, heading) in enumerate(LAYERS):
    if service != seen:
        left += SERVICE_GAP if seen is not None else 0
        titles.append({"x": left, "y": -150, "text": f"{SERVICES[service][0]} ({len(communities[service])} functions)"})
        seen = service
    subtitles.append({"x": left, "y": -80, "text": heading})
    for path, y0 in stack(order[c])[1]:
        text = short_path(path)
        headers.append({"x": left, "y": y0, "text": text if len(text) <= 38 else "…" + text[-37:]})
    for path, ids in order[c]:
        positions.update({i: (left + FN_X, ys[i]) for i in ids})
    left += COL_W
titles.append({"x": 0, "y": -230, "text": "Teal lines: HTTP (Svelte to Spring Boot) and gRPC (Spring Boot to Go)"})

assert set(positions) == set(nodes) == set(G.nodes), "every function must be placed exactly once"
assert len(set(positions.values())) == len(positions), "two functions share a position"

out = "graphify-out/graph-by-service.html"
assert to_html(G, communities, out, community_labels={i: s[0] for i, s in enumerate(SERVICES)}) is not False

html = Path(out).read_text(encoding="utf-8")
# The exporter hides low-degree labels (font size 0), but here the names are the point.
html = patch(html, "font: n.font, title: n.title,", "font: {size: 13, color: n.font.color}, title: n.title,")
html = patch(html, "x: 30 * Math.sqrt(i) * Math.cos(i * 2.4),", "x: POS[n.id][0], fixed: true,")
html = patch(html, "y: 30 * Math.sqrt(i) * Math.sin(i * 2.4),", "y: POS[n.id][1],")
html = patch(html, "const edgesDS = new vis.DataSet(", "const NETWORK = /^(http|grpc) /, NETWORK_COLOR = {color: '#14b8a6', highlight: '#2dd4bf', hover: '#2dd4bf', opacity: 0.95};\nconst edgesDS = new vis.DataSet(")
html = patch(html, "width: e.width,", "width: NETWORK.test(e.title) ? 3 : e.width,")
html = patch(html, "color: e.color,", "color: NETWORK.test(e.title) ? NETWORK_COLOR : e.color,")
html = patch(html, "// Build vis datasets", "const POS = " + json.dumps(positions) + ";")
html = patch(html, "physics: {\n    enabled: true,", "physics: {\n    enabled: false,")
html = patch(html, "const container = document.getElementById('graph');", "const container = document.getElementById('graph');\nconst TITLES = " + json.dumps(titles) + ", SUBTITLES = " + json.dumps(subtitles) + ", HEADERS = " + json.dumps(headers) + ";")
html = patch(html, "smooth: { type: 'continuous', roundness: 0.2 }", "smooth: { type: 'cubicBezier', forceDirection: 'horizontal', roundness: 0.5 }")
html = patch(html, "network.once('stabilizationIterationsDone', () => {", """network.on('afterDrawing', ctx => {
  ctx.textBaseline = 'top';
  const ink = getComputedStyle(document.body).color;
  ctx.fillStyle = ink; ctx.font = 'bold 44px sans-serif';
  TITLES.forEach(t => ctx.fillText(t.text, t.x, t.y));
  ctx.font = 'bold 24px sans-serif';
  SUBTITLES.forEach(t => ctx.fillText(t.text, t.x, t.y));
  ctx.font = '15px monospace';
  HEADERS.forEach(h => ctx.fillText(h.text, h.x, h.y));
});
network.once('stabilizationIterationsDone', () => {""")
Path(out).write_text(html, encoding="utf-8")

# Every network call must be accounted for: drawn as an edge or listed below.
print("\nFLOWS (Svelte function | route | controller | service | Go handler)")
adj = {}
for a, b, relation in network_edges:
    adj.setdefault(a, []).append((b, relation))
for a, ctrl, http in sorted(e for e in network_edges if e[2].startswith("http")):
    services = [
        (svc, *rpc) for svc in G.neighbors(ctrl)
        if G.edges[ctrl, svc]["_src"] == ctrl and "/service/" in nodes[svc]["source_file"]
        for rpc in adj.get(svc, []) if rpc[1].startswith("grpc")
    ]
    tail = " ; ".join(f'{nodes[s]["source_file"].split("/")[-1]}{nodes[s]["label"]} | {nodes[g]["label"]}' for s, g, _r in services) or "(no gRPC hop found)"
    print(f'{nodes[a]["label"]} | {http[5:]} | {nodes[ctrl]["source_file"].split("/")[-1]}{nodes[ctrl]["label"]} | {tail}')
print(f"\nurls with no Spring route: {unmatched_urls}\nroutes no UI calls: {unused_routes}\nstub calls with no Go handler: {unimplemented}\nGo handlers Spring never calls: {uncalled_rpcs}\n")

cross = sum(service_of(nodes[a]["source_file"]) != service_of(nodes[b]["source_file"]) for a, b in G.edges)
for c, (service, heading) in enumerate(LAYERS):
    print(f"{SERVICES[service][0]} / {heading}: {sum(len(ids) for _p, ids in order[c])} functions in {len(order[c])} files")
print(f"{G.number_of_edges()} edges, {cross} cross-service, {sum(1 for n in G if G.degree(n) == 0)} functions with no edges")

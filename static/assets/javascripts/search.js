const onLoad = async () => {
  const GRAM_N = 2;
  const toHexString = (n) => {
    return n.toString(16).padStart(2, "0");
  };
  const to2dString = (n) => {
    return n.toString().padStart(2, "0");
  };

  class Uint8ArrayReader {
    constructor(u8ary) {
      this.u8ary = u8ary;
      this.i = 0;
    }

    get length() {
      return this.u8ary.length;
    }

    readInt() {
      const n = (this.u8ary[this.i] << 24) | (this.u8ary[this.i + 1] << 16) | (this.u8ary[this.i + 2] << 8) | this.u8ary[this.i + 3];
      this.i += 4;
      return n;
    }

    readVInt() {
      let n = 0;
      let cont = true;
      for (; cont; this.i++) {
        n <<= 7;
        n |= this.u8ary[this.i] & 0b01111111;
        cont = (this.u8ary[this.i] & 0b10000000) != 0;
      }
      return n;
    }

    isEOF() {
      return this.u8ary.length <= this.i;
    }
  }

  class Index {
    constructor() {
      this.indexes = new Map();
    }

    async get(key) {
      const cached = this.indexes.get(key);
      if (cached) {
        return cached;
      }
      const paths = [];
      for (let i = 0; i < key.length; i++) {
        const n = key.charCodeAt(i);
        paths.push(toHexString(n >> 8));
        paths.push(toHexString(n & 0xff));
      }
      const res = await fetch(`./index/${paths.join("/")}.index`);
      const index = new Map();
      if (res.ok) {
        const blob = await res.blob();
        const reader = new Uint8ArrayReader(new Uint8Array(await blob.arrayBuffer()));
        while (!reader.isEOF()) {
          const channelNumber = reader.readVInt();
          let mesCount = reader.readVInt();
          while (0 <= --mesCount) {
            const tsSec = reader.readInt();
            const tsMicrosec = reader.readVInt();
            const ts = `${tsSec}.${tsMicrosec.toString().padStart(6, "0")}`;

            const key = `${channelNumber}:${ts}`;
            let posSet = index.get(key);
            if (posSet == null) {
              posSet = new Set();
              index.set(key, posSet);
            }
            for (;;) {
              const pos = reader.readVInt();
              if (pos === 0) {
                break;
              }
              posSet.add(pos - 1);
            }
          }
        }
      } else {
        // TODO: check error type
      }
      this.indexes.set(key, index);
      return index;
    }
  }

  const index = new Index();
  const sepRegexp = new RegExp(`.{1,${GRAM_N}}`, "g");

  const numToChannel = await (async () => {
    const map = new Map();
    const res = await fetch("./index/channel");
    for (const line of (await res.text()).split("\n")) {
      const [n, channelID, channelName] = line.split("\t");
      if (channelID != null) {
        map.set(n - 0, {channelID, channelName});
      }
    }
    return map;
  })();

  const searchByWord = async (word) => {
    const indexes = await Promise.all(
      [...word.matchAll(sepRegexp)].map(async (chunk) => index.get(chunk[0]))
    );
    const result = indexes.reduce((acc, cur) => {
      const next = new Map();
      for (const [key, prevPosSet] of acc.entries()) {
        const currentPosSet = cur.get(key);
        for (const prevPos of prevPosSet.values()) {
          if (currentPosSet?.has(prevPos + GRAM_N)) {
            let nestPosSet = next.get(key);
            if (nestPosSet == null) {
              nestPosSet = new Set();
              next.set(key, nestPosSet);
            }
            nestPosSet.add(prevPos + GRAM_N);
          }
        }
      }
      return next;
    });

    return result;
  };

  const parseQuery = (query) => {
    // TODO: parse
    return query;
  };

  const search = async (query) => {
    const word = parseQuery(query);
    return searchByWord(word);
  };

  const text = document.getElementById("search-text");
  const resultElement = document.getElementById("result");

  const execute = async () => {
    try {
      const startTime = Date.now();

      const result = await search(text.value);

      const links = [...result.keys()].map((doc) => {
        const [channelNumber, ts] = doc.split(":");
        const {channelID, channelName} = numToChannel.get(channelNumber - 0);
        const date = new Date(ts.split(".")[0] * 1000);
        const link = `${channelID}/${date.getFullYear()}/${(date.getMonth() + 1).toString().padStart(2, "0")}/#ts-${ts}`;
        return `<a href="${link}">&#35;${channelName}: ${date.getFullYear()}-${to2dString(date.getMonth() + 1)}-${to2dString(date.getDay())} ${to2dString(date.getHours())}:${to2dString(date.getMinutes())}:${to2dString(date.getSeconds())}</a>`;
      });
      const processTime = Date.now() - startTime;
      resultElement.innerHTML = `<p>${links.length} 件ヒットしました (${processTime / 1000} 秒)</p>${links.join("<br>")}`;
    } catch (e) {
      resultElement.innerHTML = `検索中にエラーが発生しました: ${e.name}: ${e.message}`;
    }
  };

  const searchFromParams = () => {
    const query = new URL(location).searchParams.get("q");
    if (query) {
      text.value = query;
      execute();
    } else {
      text.value = "";
      resultElement.innerHTML = "";
    }
  };

  const moveToResultPage = () => {
    const params = new URLSearchParams();
    params.set("q", text.value);
    history.pushState({q: text.value}, document.title, `${location.pathname}?${params.toString()}`);
    execute();
  };

  document.getElementById("search-button").addEventListener("click", moveToResultPage);
  document.getElementById("search-text").addEventListener("keypress", (event) => {
    if (event.key === "Enter") {
      moveToResultPage();
    }
  });
  window.addEventListener("popstate", (e) => {
    console.log(e);
    searchFromParams();
  });

  searchFromParams();
};

window.addEventListener('DOMContentLoaded', () => {
  onLoad();
});

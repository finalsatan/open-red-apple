export namespace dto {
	
	export class Season {
	    seriesId: string;
	    title: string;
	    index: number;
	    coverUrl: string;
	    episodeCount: number;
	    status: number;
	
	    static createFrom(source: any = {}) {
	        return new Season(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seriesId = source["seriesId"];
	        this.title = source["title"];
	        this.index = source["index"];
	        this.coverUrl = source["coverUrl"];
	        this.episodeCount = source["episodeCount"];
	        this.status = source["status"];
	    }
	}
	export class Album {
	    id: string;
	    title: string;
	    description: string;
	    coverUrl: string;
	    seasonCount: number;
	    seasons: Season[];
	
	    static createFrom(source: any = {}) {
	        return new Album(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.coverUrl = source["coverUrl"];
	        this.seasonCount = source["seasonCount"];
	        this.seasons = this.convertValues(source["seasons"], Season);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Episode {
	    id: string;
	    number: number;
	    title: string;
	    playable: boolean;
	    vertical: boolean;
	    isNew: boolean;
	    coverUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new Episode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.number = source["number"];
	        this.title = source["title"];
	        this.playable = source["playable"];
	        this.vertical = source["vertical"];
	        this.isNew = source["isNew"];
	        this.coverUrl = source["coverUrl"];
	    }
	}
	export class EpisodePage {
	    items: Episode[];
	    cursor: string;
	    hasMore: boolean;
	
	    static createFrom(source: any = {}) {
	        return new EpisodePage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], Episode);
	        this.cursor = source["cursor"];
	        this.hasMore = source["hasMore"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Favorite {
	    seriesId: string;
	    title: string;
	    coverUrl: string;
	    addedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new Favorite(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seriesId = source["seriesId"];
	        this.title = source["title"];
	        this.coverUrl = source["coverUrl"];
	        this.addedAt = source["addedAt"];
	    }
	}
	export class FilterOption {
	    name: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new FilterOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.value = source["value"];
	    }
	}
	export class FilterRow {
	    key: string;
	    name: string;
	    options: FilterOption[];
	
	    static createFrom(source: any = {}) {
	        return new FilterRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.name = source["name"];
	        this.options = this.convertValues(source["options"], FilterOption);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class HistoryItem {
	    seriesId: string;
	    episodeId: string;
	    episodeNumber: number;
	    title: string;
	    coverUrl: string;
	    position: number;
	    duration: number;
	    finished: boolean;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new HistoryItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seriesId = source["seriesId"];
	        this.episodeId = source["episodeId"];
	        this.episodeNumber = source["episodeNumber"];
	        this.title = source["title"];
	        this.coverUrl = source["coverUrl"];
	        this.position = source["position"];
	        this.duration = source["duration"];
	        this.finished = source["finished"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class HotTopic {
	    id: string;
	    title: string;
	    query: string;
	    heat: string;
	
	    static createFrom(source: any = {}) {
	        return new HotTopic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.query = source["query"];
	        this.heat = source["heat"];
	    }
	}
	export class HotItem {
	    position: number;
	    seriesId: string;
	    title: string;
	    coverUrl: string;
	    episodeCount: number;
	    searchScore: string;
	    heat: string;
	    tag: string;
	
	    static createFrom(source: any = {}) {
	        return new HotItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.position = source["position"];
	        this.seriesId = source["seriesId"];
	        this.title = source["title"];
	        this.coverUrl = source["coverUrl"];
	        this.episodeCount = source["episodeCount"];
	        this.searchScore = source["searchScore"];
	        this.heat = source["heat"];
	        this.tag = source["tag"];
	    }
	}
	export class HotBoard {
	    name: string;
	    items: HotItem[];
	    topics: HotTopic[];
	
	    static createFrom(source: any = {}) {
	        return new HotBoard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.items = this.convertValues(source["items"], HotItem);
	        this.topics = this.convertValues(source["topics"], HotTopic);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class HotSearch {
	    guesses: HotItem[];
	    boards: HotBoard[];
	
	    static createFrom(source: any = {}) {
	        return new HotSearch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.guesses = this.convertValues(source["guesses"], HotItem);
	        this.boards = this.convertValues(source["boards"], HotBoard);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class InfoRow {
	    kind: number;
	    text: string;
	
	    static createFrom(source: any = {}) {
	        return new InfoRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.text = source["text"];
	    }
	}
	export class NewRelease {
	    seriesId: string;
	    title: string;
	    description: string;
	    coverUrl: string;
	    episodeCount: number;
	    online: boolean;
	    publishAt: string;
	    categories: string[];
	    heat: string;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new NewRelease(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seriesId = source["seriesId"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.coverUrl = source["coverUrl"];
	        this.episodeCount = source["episodeCount"];
	        this.online = source["online"];
	        this.publishAt = source["publishAt"];
	        this.categories = source["categories"];
	        this.heat = source["heat"];
	        this.note = source["note"];
	    }
	}
	export class NewReleasePage {
	    date: string;
	    dates: string[];
	    items: NewRelease[];
	    hasMore: boolean;
	    cursor: string;
	
	    static createFrom(source: any = {}) {
	        return new NewReleasePage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.dates = source["dates"];
	        this.items = this.convertValues(source["items"], NewRelease);
	        this.hasMore = source["hasMore"];
	        this.cursor = source["cursor"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Person {
	    userId: string;
	    name: string;
	    description: string;
	    avatarUrl: string;
	    tags: string[];
	    labels: string[];
	    followers: number;
	    likes: number;
	
	    static createFrom(source: any = {}) {
	        return new Person(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.userId = source["userId"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.avatarUrl = source["avatarUrl"];
	        this.tags = source["tags"];
	        this.labels = source["labels"];
	        this.followers = source["followers"];
	        this.likes = source["likes"];
	    }
	}
	export class RankBoard {
	    id: string;
	    name: string;
	    selected: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RankBoard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.selected = source["selected"];
	    }
	}
	export class RankChange {
	    kind: number;
	    places: number;
	
	    static createFrom(source: any = {}) {
	        return new RankChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.places = source["places"];
	    }
	}
	export class RankContent {
	    id: string;
	    name: string;
	    selected: boolean;
	    boards: RankBoard[];
	
	    static createFrom(source: any = {}) {
	        return new RankContent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.selected = source["selected"];
	        this.boards = this.convertValues(source["boards"], RankBoard);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RankItem {
	    rank: number;
	    position: number;
	    seriesId: string;
	    title: string;
	    coverUrl: string;
	    score: string;
	    episodeCount: number;
	    promoted: boolean;
	    actor: string;
	    userId: string;
	    avatarUrl: string;
	    intro: string;
	    details: string[];
	    change?: RankChange;
	
	    static createFrom(source: any = {}) {
	        return new RankItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rank = source["rank"];
	        this.position = source["position"];
	        this.seriesId = source["seriesId"];
	        this.title = source["title"];
	        this.coverUrl = source["coverUrl"];
	        this.score = source["score"];
	        this.episodeCount = source["episodeCount"];
	        this.promoted = source["promoted"];
	        this.actor = source["actor"];
	        this.userId = source["userId"];
	        this.avatarUrl = source["avatarUrl"];
	        this.intro = source["intro"];
	        this.details = source["details"];
	        this.change = this.convertValues(source["change"], RankChange);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Series {
	    id: string;
	    title: string;
	    description: string;
	    coverUrl: string;
	    episodeCount: number;
	    status: number;
	    payType: number;
	    contentType: number;
	    platform: number;
	    heat: number;
	    playCount: number;
	    followCount: number;
	    diggCount: number;
	    rightText: string;
	    subtitles: string[];
	    isMultiSeason: boolean;
	    seasonIndex: number;
	    rows: InfoRow[];
	
	    static createFrom(source: any = {}) {
	        return new Series(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.coverUrl = source["coverUrl"];
	        this.episodeCount = source["episodeCount"];
	        this.status = source["status"];
	        this.payType = source["payType"];
	        this.contentType = source["contentType"];
	        this.platform = source["platform"];
	        this.heat = source["heat"];
	        this.playCount = source["playCount"];
	        this.followCount = source["followCount"];
	        this.diggCount = source["diggCount"];
	        this.rightText = source["rightText"];
	        this.subtitles = source["subtitles"];
	        this.isMultiSeason = source["isMultiSeason"];
	        this.seasonIndex = source["seasonIndex"];
	        this.rows = this.convertValues(source["rows"], InfoRow);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Recommend {
	    cellTitle: string;
	    items: Series[];
	
	    static createFrom(source: any = {}) {
	        return new Recommend(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cellTitle = source["cellTitle"];
	        this.items = this.convertValues(source["items"], Series);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class SeriesPage {
	    items: Series[];
	    cursor: string;
	    hasMore: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SeriesPage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], Series);
	        this.cursor = source["cursor"];
	        this.hasMore = source["hasMore"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ShortFilterPage {
	    items: Series[];
	    hasMore: boolean;
	    cursor: string;
	
	    static createFrom(source: any = {}) {
	        return new ShortFilterPage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], Series);
	        this.hasMore = source["hasMore"];
	        this.cursor = source["cursor"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ShortRanking {
	    contents: RankContent[];
	    content: string;
	    board: string;
	    items: RankItem[];
	    hasMore: boolean;
	    cursor: string;
	
	    static createFrom(source: any = {}) {
	        return new ShortRanking(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.contents = this.convertValues(source["contents"], RankContent);
	        this.content = source["content"];
	        this.board = source["board"];
	        this.items = this.convertValues(source["items"], RankItem);
	        this.hasMore = source["hasMore"];
	        this.cursor = source["cursor"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Suggestion {
	    text: string;
	    keyword: string;
	    kind: number;
	
	    static createFrom(source: any = {}) {
	        return new Suggestion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text = source["text"];
	        this.keyword = source["keyword"];
	        this.kind = source["kind"];
	    }
	}
	export class Work {
	    seriesId: string;
	    title: string;
	    description: string;
	    coverUrl: string;
	    episodeCount: number;
	    tags: string[];
	    heat: number;
	    plays: number;
	
	    static createFrom(source: any = {}) {
	        return new Work(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seriesId = source["seriesId"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.coverUrl = source["coverUrl"];
	        this.episodeCount = source["episodeCount"];
	        this.tags = source["tags"];
	        this.heat = source["heat"];
	        this.plays = source["plays"];
	    }
	}
	export class WorkPage {
	    items: Work[];
	    hasMore: boolean;
	    cursor: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkPage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], Work);
	        this.hasMore = source["hasMore"];
	        this.cursor = source["cursor"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}


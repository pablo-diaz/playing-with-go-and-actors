import http from 'k6/http';
import { vu, scenario } from 'k6/execution';

const albumIds = ["al01", "al02", "al03"];

function getAlbumIdToUse(forTestId) {
    const index = (forTestId - 1) % albumIds.length;
    return albumIds[index];
}

export default function () {
    http.get(`http://localhost:8080/albums/new/${getAlbumIdToUse(vu.idInTest)}?s=${scenario.name}`);
}
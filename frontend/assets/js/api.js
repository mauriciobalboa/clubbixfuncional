/* ==========================================================================
   CLUBBIX — api.js
   Helper centralizado para todas as chamadas ao backend Go.
   Lê o token JWT diretamente do clubbix_session (localStorage/sessionStorage).
   ========================================================================== */
var ClubAPI = (function () {
  'use strict';

  /* O backend serve o frontend no mesmo origem, então BASE fica vazio.       */
  /* Se precisar apontar para outro host, mude aqui: 'http://localhost:8080'  */
  var BASE = '';

  /* ── leitura do token armazenado na sessão ────────────────────────────── */
  function getToken() {
    try {
      var raw = localStorage.getItem('clubbix_session') || sessionStorage.getItem('clubbix_session');
      if (!raw) return null;
      var s = JSON.parse(raw);
      return (s && s.token) ? s.token : null;
    } catch (e) { return null; }
  }

  /* ── monta headers com Content-Type + Authorization (se autenticado) ──── */
  function buildHeaders() {
    var h = { 'Content-Type': 'application/json' };
    var tok = getToken();
    if (tok) h['Authorization'] = 'Bearer ' + tok;
    return h;
  }

  /* ── resposta padronizada: rejeita se !r.ok ────────────────────────────── */
  function handleResponse(r) {
    return r.json().then(function (data) {
      if (!r.ok) return Promise.reject(data);
      return data;
    });
  }

  /* ── GET autenticado ───────────────────────────────────────────────────── */
  function get(path) {
    return fetch(BASE + path, { headers: buildHeaders() }).then(handleResponse);
  }

  /* ── POST com body JSON ────────────────────────────────────────────────── */
  function post(path, body) {
    return fetch(BASE + path, {
      method: 'POST',
      headers: buildHeaders(),
      body: JSON.stringify(body || {})
    }).then(handleResponse);
  }

  /* ── PUT com body JSON ─────────────────────────────────────────────────── */
  function put(path, body) {
    return fetch(BASE + path, {
      method: 'PUT',
      headers: buildHeaders(),
      body: JSON.stringify(body || {})
    }).then(handleResponse);
  }

  /* ── DELETE ────────────────────────────────────────────────────────────── */
  function del(path) {
    return fetch(BASE + path, {
      method: 'DELETE',
      headers: buildHeaders()
    }).then(handleResponse);
  }

  return { get: get, post: post, put: put, del: del, getToken: getToken };
})();

/* ========================================================================
   CLUBBIX — painel · lógica de login, clube e assinatura (mock front-end)
   ======================================================================== */
(function () {
  'use strict';

  var $ = function (s, c) { return (c || document).querySelector(s); };
  var $$ = function (s, c) { return Array.prototype.slice.call((c || document).querySelectorAll(s)); };
  var body = document.body;
  var reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  var WHATS = 'https://wa.me/5565992527948';

  /* ─── utilitários: throttle · btnBusy ─────────────────────────────── */
  function throttle(fn, limit) {
    var last = 0;
    return function () {
      var now = Date.now();
      if (now - last < limit) return;
      last = now;
      return fn.apply(this, arguments);
    };
  }
  function btnBusy(el, html) {
    if (!el) return;
    el.disabled = true;
    el.dataset.origHtml = el.innerHTML;
    if (html) el.innerHTML = html;
    el.style.opacity = '0.6';
    el.style.cursor  = 'not-allowed';
  }
  function btnFree(el) {
    if (!el) return;
    el.disabled = false;
    if (el.dataset.origHtml) el.innerHTML = el.dataset.origHtml;
    el.style.opacity = '';
    el.style.cursor  = '';
  }

  /* ---------------------- reveal ao surgir ----------------------------- */
  (function () {
    var els = $$('.reveal');
    if (!('IntersectionObserver' in window) || reduced) {
      els.forEach(function (el) { el.classList.add('is-visible'); });
      return;
    }
    var obs = new IntersectionObserver(function (entries) {
      entries.forEach(function (e) {
        if (e.isIntersecting) { e.target.classList.add('is-visible'); obs.unobserve(e.target); }
      });
    }, { threshold: 0.12, rootMargin: '0px 0px -6% 0px' });
    els.forEach(function (el) { obs.observe(el); });
  })();

  /* ---------------------- sessão (localStorage) ------------------------ */
  var KEY = 'clubbix_session';
  function getSession() {
    try { return JSON.parse(localStorage.getItem(KEY) || sessionStorage.getItem(KEY)); }
    catch (e) { return null; }
  }
  function setSession(obj, persist) {
    var json = JSON.stringify(obj);
    if (persist === false) { sessionStorage.setItem(KEY, json); localStorage.removeItem(KEY); }
    else { localStorage.setItem(KEY, json); sessionStorage.removeItem(KEY); }
  }
  function clearSession() { localStorage.removeItem(KEY); sessionStorage.removeItem(KEY); }
  function firstName(n) { return (n || '').trim().split(/\s+/)[0] || 'Membro'; }
  function initials(n) {
    var p = (n || 'C').trim().split(/\s+/);
    return ((p[0] || 'C')[0] + (p.length > 1 ? p[p.length - 1][0] : '')).toUpperCase();
  }
  function genId() { return 'CLB-2026-' + String(Math.floor(1000 + Math.random() * 9000)); }

  /* máscaras */
  function digits(s) { return s.replace(/\D/g, ''); }
  function maskCPF(v) {
    v = digits(v).slice(0, 11);
    if (v.length > 9) return v.replace(/(\d{3})(\d{3})(\d{3})(\d{1,2})/, '$1.$2.$3-$4');
    if (v.length > 6) return v.replace(/(\d{3})(\d{3})(\d{1,3})/, '$1.$2.$3');
    if (v.length > 3) return v.replace(/(\d{3})(\d{1,3})/, '$1.$2');
    return v;
  }
  function maskPhone(v) {
    v = digits(v).slice(0, 11);
    if (v.length > 10) return v.replace(/(\d{2})(\d{5})(\d{1,4})/, '($1) $2-$3');
    if (v.length > 6) return v.replace(/(\d{2})(\d{4,5})(\d{1,4})/, '($1) $2-$3');
    if (v.length > 2) return v.replace(/(\d{2})(\d{1,5})/, '($1) $2');
    if (v.length > 0) return '(' + v;
    return v;
  }
  function maskCard(v) { return digits(v).slice(0, 16).replace(/(\d{4})(?=\d)/g, '$1 '); }
  function maskExp(v) {
    v = digits(v).slice(0, 4);
    if (v.length > 2) return v.replace(/(\d{2})(\d{1,2})/, '$1/$2');
    return v;
  }
  function bindMask(el, fn) {
    if (!el) return;
    el.addEventListener('input', function () {
      var p = el.selectionStart, len = el.value.length;
      el.value = fn(el.value);
      if (p === len) el.setSelectionRange(el.value.length, el.value.length);
    });
  }
  var EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

  /* ====================== PÁGINA: LOGIN ================================ */
  function initLogin() {
    var form = $('#loginForm');
    var msg = $('#loginMsg');

    var toggle = $('#passToggle'), pass = $('#loginPass');
    toggle.addEventListener('click', function () {
      var show = pass.type === 'password';
      pass.type = show ? 'text' : 'password';
      toggle.querySelector('use').setAttribute('href', show ? '#i-eye-off' : '#i-eye');
      toggle.setAttribute('aria-label', show ? 'Ocultar senha' : 'Mostrar senha');
    });

    $('#forgotBtn').addEventListener('click', function () {
      msg.className = 'auth-msg';
      msg.textContent = 'Para recuperar a senha, fale com a gente no WhatsApp (65) 99252-7948. 💬';
    });

    var _loginBtn = form.querySelector('[type="submit"]');
    form.addEventListener('submit', function (e) {
      e.preventDefault();
      btnBusy(_loginBtn, 'Entrando...');
      var email = $('#loginEmail').value.trim().toLowerCase();
      var senha = $('#loginPass').value;
      if (!email || !senha) {
        msg.className = 'auth-msg is-error';
        msg.textContent = 'Preencha o e-mail e a senha. 🙂';
        btnFree(_loginBtn);
        return;
      }
      var persist = $('#remember').checked;
      ClubAPI.post('/api/auth/login', { email: email, senha: senha })
        .then(function (data) {
          var session = data.session || {};
          if (!session.token && data.token) session.token = data.token;
          setSession(session, persist);
          msg.className = 'auth-msg is-ok';
          msg.textContent = 'Tudo certo! Entrando no seu clube... 🎉';
          setTimeout(function () { window.location.href = 'clube.html'; }, reduced ? 150 : 700);
        })
        .catch(function (err) {
          msg.className = 'auth-msg is-error';
          msg.textContent = (err && err.erro) ? err.erro : 'E-mail ou senha incorretos.';
          btnFree(_loginBtn);
        });
    });
  }

  /* ====================== PÁGINA: CLUBE ================================ */
  /* Fallback de ícones/nomes caso a API não retorne categoria_icone */
  var CATS = {
    restaurantes: { l: 'Restaurantes', e: '🍔' }, saude: { l: 'Saúde', e: '🩺' },
    educacao: { l: 'Educação', e: '📚' }, beleza: { l: 'Beleza', e: '💇' },
    academias: { l: 'Academias', e: '🏋️' }, moda: { l: 'Moda', e: '🛍️' },
    lazer: { l: 'Lazer', e: '🎬' }, farmacias: { l: 'Farmácias', e: '💊' },
    pet: { l: 'Pet', e: '🐾' }, turismo: { l: 'Turismo', e: '✈️' },
    automotivo: { l: 'Automotivo', e: '🚗' }, servicos: { l: 'Serviços', e: '🔧' }
  };

  function norm(s) { return s.toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, ''); }

  function initClube() {
    var session = getSession();
    if (!session || !session.token) { window.location.replace('login.html'); return; }

    var yearEl = $('#year'); if (yearEl) yearEl.textContent = new Date().getFullYear();

    /* ── dados do membro (atualiza via API) ──────────────────────────── */
    function populateMember(s) {
      $('#memberName').textContent = firstName(s.nome);
      $('#userNameTop').textContent = firstName(s.nome);
      $('#userAvatar').textContent = initials(s.nome);
      $('#mcName').textContent = s.nome || 'Membro Clubbix';
      $('#mcPlan').textContent = s.plano || 'Plano Clubbix';
      $('#mcId').textContent = s.memberId || s.member_id || genId();
    }
    populateMember(session);

    /* atualiza sessão com dados frescos do servidor */
    ClubAPI.get('/api/auth/me')
      .then(function (usuario) {
        var s = {
          nome:     usuario.nome_completo || session.nome,
          email:    usuario.email        || session.email,
          token:    session.token,
          memberId: 'CLB-2026-' + usuario.id_usuario
        };
        if (usuario.assinaturas && usuario.assinaturas.length > 0) {
          var ult = usuario.assinaturas[0];
          if (ult.plano && ult.plano.nome_plano) s.plano = 'Plano ' + ult.plano.nome_plano;
        }
        s.plano = s.plano || session.plano || 'Plano Clubbix';
        setSession(s, true);
        populateMember(s);
      })
      .catch(function () {
        /* token expirado ou inválido → redireciona para login */
        clearSession();
        window.location.replace('login.html');
      });

    $('#logoutBtn').addEventListener('click', function () {
      ClubAPI.post('/api/auth/logout', {}).catch(function(){});
      clearSession();
      window.location.href = 'index.html';
    });

    /* ── diretório de parceiros (dados reais da API) ─────────────────── */
    var grid    = $('#partnerGrid');
    var empty   = $('#partnerEmpty');
    var count   = $('#discCount');
    var search  = $('#partnerSearch');
    var state   = { cat: 'todos', q: '' };
    var allPartners = [];

    function renderPartners() {
      var list = allPartners.filter(function (p) {
        var okCat = state.cat === 'todos' || p.categoria === state.cat;
        var okQ   = !state.q || norm(p.nome_empresa || '').indexOf(state.q) !== -1;
        return okCat && okQ;
      });
      grid.innerHTML = '';
      if (list.length === 0) {
        empty.hidden = false;
        count.textContent = '0 parceiros encontrados';
        return;
      }
      empty.hidden = true;
      list.forEach(function (p) {
        var catFallback = CATS[p.categoria] || { l: p.categoria_nome || p.categoria, e: p.categoria_icone || '🏪' };
        var ico  = p.categoria_icone || catFallback.e;
        var nome = p.categoria_nome  || catFallback.l;
        var disc = p.percentual_desconto > 0 ? p.percentual_desconto + '% OFF' : 'Desconto exclusivo';
        var card = document.createElement('button');
        card.type = 'button';
        card.className = 'partner-card';
        card.innerHTML =
          '<div class="pc-top"><span class="pc-emoji">' + ico + '</span>' +
          '<span class="pc-disc">' + disc + '</span></div>' +
          '<span class="pc-name">' + p.nome_empresa + '</span>' +
          '<span class="pc-cat">' + nome + '</span>' +
          '<span class="pc-loc"><svg class="ico"><use href="#i-pin"/></svg>' + (p.endereco || '') + '</span>' +
          '<span class="pc-go">Ver desconto <svg class="ico"><use href="#i-arrow"/></svg></span>';
        card.addEventListener('click', function () { openModal(p); });
        grid.appendChild(card);
      });
      count.textContent = list.length + (list.length === 1 ? ' parceiro encontrado' : ' parceiros encontrados');
    }

    /* carrega parceiros da API */
    if (grid) {
      grid.innerHTML = '<p style="grid-column:1/-1;text-align:center;opacity:.5">Carregando parceiros...</p>';
      ClubAPI.get('/api/parceiros')
        .then(function (data) {
          allPartners = data.parceiros || [];
          renderPartners();
        })
        .catch(function () {
          grid.innerHTML = '<p style="grid-column:1/-1;text-align:center;opacity:.5">Não foi possível carregar os parceiros.</p>';
        });
    }

    $$('.catpill').forEach(function (pill) {
      pill.addEventListener('click', function () {
        $$('.catpill').forEach(function (p) { p.classList.remove('is-active'); });
        pill.classList.add('is-active');
        state.cat = pill.dataset.cat;
        renderPartners();
      });
    });
    if (search) search.addEventListener('input', function () { state.q = norm(search.value.trim()); renderPartners(); });

    /* modal de parceiro */
    var modal = $('#partnerModal');
    function openModal(p) {
      var catFallback = CATS[p.categoria] || { l: p.categoria_nome || p.categoria, e: p.categoria_icone || '🏪' };
      var disc = p.percentual_desconto > 0 ? p.percentual_desconto + '% OFF' : 'Desconto exclusivo';
      $('#pmEmoji').textContent = p.categoria_icone || catFallback.e;
      $('#pmCat').textContent   = p.categoria_nome  || catFallback.l;
      $('#pmName').textContent  = p.nome_empresa;
      $('#pmDisc').textContent  = disc;
      $('#pmLoc').textContent   = (p.endereco || '') + (p.endereco ? ' — Cuiabá-MT' : 'Cuiabá-MT');
      $('#pmCode').textContent  = p.codigo || 'CLB-' + p.id_parceiro;
      modal.classList.add('is-open');
      modal.setAttribute('aria-hidden', 'false');
    }
    function closeModal() {
      modal.classList.remove('is-open');
      modal.setAttribute('aria-hidden', 'true');
    }
    $$('[data-close-modal]', modal).forEach(function (el) { el.addEventListener('click', closeModal); });
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape' && modal.classList.contains('is-open')) closeModal();
    });
  }

  /* ====================== PÁGINA: ASSINAR ============================== */
  function initAssinar() {
    var form = $('#assinarForm');
    var msg = $('#assinarMsg');
    var sumPlan = $('#sumPlan'), sumPrice = $('#sumPrice');

    /* ── código de convite via ?ref= ──────────────────────────────────── */
    (function () {
      var params  = new URLSearchParams(window.location.search);
      var refCode = params.get('ref');
      if (!refCode) return;

      /* gera um código de convite aleatório no formato INV-XXXX-DDDD */
      var chars   = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789'; // sem ambíguos (I,O,0,1)
      var letras  = Array.from({ length: 4 }, function () {
        return chars[Math.floor(Math.random() * chars.length)];
      }).join('');
      var digitos = Math.floor(1000 + Math.random() * 9000);
      var invCode = 'INV-' + letras + '-' + digitos;

      var banner  = $('#inviteBanner');
      var field   = $('#inviteField');
      var input   = $('#aConvite');

      if (banner) banner.hidden = false;
      if (field)  field.hidden  = false;
      if (input)  input.value   = invCode;

      /* anima a entrada do banner */
      if (banner && !window.matchMedia('(prefers-reduced-motion:reduce)').matches) {
        banner.style.animation = 'heroFadeUp .55s cubic-bezier(.16,1,.3,1) both';
      }
    })();

    /* pré-seleção via ?plano= */
    var planos = $$('input[name="plano"]');
    var params2 = new URLSearchParams(window.location.search);
    var preset  = params2.get('plano');
    if (preset) {
      planos.forEach(function (r) { if (r.value === preset) r.checked = true; });
    }
    function updateSummary() {
      var sel = planos.filter(function (r) { return r.checked; })[0];
      if (sel) {
        sumPlan.textContent = sel.dataset.label + ' · ' + sel.dataset.dur;
        sumPrice.textContent = sel.dataset.price;
      } else {
        sumPlan.textContent = 'Selecione um plano acima';
        sumPrice.textContent = '—';
      }
    }
    planos.forEach(function (r) { r.addEventListener('change', updateSummary); });
    updateSummary();

    /* método de pagamento */
    var payCard = $('#payCard'), payPix = $('#payPix');
    /* init: usa style.display para não perder para regras CSS de maior especificidade */
    payPix.style.display = 'none';
    $$('input[name="pagamento"]').forEach(function (r) {
      r.addEventListener('change', function () {
        if (!r.checked) return;
        var pix = r.value === 'pix';
        payCard.style.display = pix ? 'none' : '';   /* '' restaura o flex do CSS */
        payPix.style.display  = pix ? 'block' : 'none';
      });
    });

    /* máscaras */
    bindMask($('#aCpf'), maskCPF);
    bindMask($('#aTel'), maskPhone);
    bindMask($('#cardNum'), maskCard);
    bindMask($('#cardExp'), maskExp);
    bindMask($('#cardCvv'), function (v) { return digits(v).slice(0, 4); });

    var _assinarBtn = form.querySelector('[type="submit"]');
    var _assinarHandler = throttle(function () {
      var sel = planos.filter(function (r) { return r.checked; })[0];
      var nome = $('#aNome').value.trim();
      var email = $('#aEmail').value.trim();
      var cpf = digits($('#aCpf').value);
      var tel = digits($('#aTel').value);
      var metodo = ($$('input[name="pagamento"]').filter(function (r) { return r.checked; })[0] || {}).value;

      function fail(text, focusEl) {
        msg.className = 'auth-msg is-error';
        msg.textContent = text;
        if (focusEl) focusEl.focus();
        btnFree(_assinarBtn);
      }
      if (!sel) return fail('Escolha um plano para continuar. 👆');
      if (nome.split(/\s+/).length < 2) return fail('Digite seu nome completo.', $('#aNome'));
      if (!EMAIL_RE.test(email)) return fail('Digite um e-mail válido.', $('#aEmail'));
      if (cpf.length !== 11) return fail('Digite um CPF válido (11 dígitos).', $('#aCpf'));
      if (tel.length < 10) return fail('Digite um WhatsApp válido com DDD.', $('#aTel'));
      if (metodo === 'cartao') {
        if (digits($('#cardNum').value).length < 16) return fail('Confira o número do cartão.', $('#cardNum'));
        if (!$('#cardName').value.trim()) return fail('Digite o nome impresso no cartão.', $('#cardName'));
        if ($('#cardExp').value.length < 5) return fail('Confira a validade do cartão.', $('#cardExp'));
        if (digits($('#cardCvv').value).length < 3) return fail('Confira o CVV do cartão.', $('#cardCvv'));
      }

      msg.className = 'auth-msg';
      msg.textContent = '';

      /* ── Chamada real à API ─────────────────────────────────────────── */
      var payload = {
        nome:     nome,
        email:    email,
        cpf:      cpf,
        telefone: tel,
        plano:    sel.value,
        pagamento: metodo || 'pix',
        renovacao_automatica: false
      };

      ClubAPI.post('/api/assinaturas/publica', payload)
        .then(function (data) {
          /* salva sessão com token retornado pela API */
          var session = data.session || {};
          if (!session.token && data.token) session.token = data.token;
          if (!session.nome)  session.nome  = nome;
          if (!session.email) session.email = email;
          if (!session.plano) session.plano = sel.dataset.label;
          setSession(session, true);

          $('#successName').textContent = firstName(nome);
          $('#successPlan').textContent = (session.plano || sel.dataset.label) + ' ativo';
          var overlay = $('#successOverlay');
          overlay.classList.add('is-open');
          overlay.setAttribute('aria-hidden', 'false');
          document.body.style.overflow = 'hidden';
          btnFree(_assinarBtn);
        })
        .catch(function (err) {
          msg.className = 'auth-msg is-error';
          msg.textContent = (err && err.erro) ? err.erro : 'Erro ao processar. Tente novamente.';
          btnFree(_assinarBtn);
        });
    }, 3000);

    form.addEventListener('submit', function (e) {
      e.preventDefault();
      btnBusy(_assinarBtn, 'Processando... <svg class="ico" style="width:1em;height:1em"><use href="#i-arrow"/></svg>');
      _assinarHandler();
    });
  }

  /* ----------------------- roteamento por página ----------------------- */
  if (body.classList.contains('page-login')) initLogin();
  if (body.classList.contains('page-clube')) initClube();
  if (body.classList.contains('page-assinar')) initAssinar();

})();

/* ----------------------- tema claro/escuro --------------------------- */
(function () {
  'use strict';
  var root = document.documentElement;
  var meta = document.querySelector('meta[name="theme-color"]');
  function paint() {
    var dark = root.getAttribute('data-theme') === 'dark';
    if (meta) meta.setAttribute('content', dark ? '#160a28' : '#fdf8f3');
  }
  paint();
  var btn = document.getElementById('themeToggle');
  if (btn) btn.addEventListener('click', function () {
    var dark = root.getAttribute('data-theme') === 'dark';
    root.setAttribute('data-theme', dark ? 'light' : 'dark');
    try { localStorage.setItem('clubbix_theme', dark ? 'light' : 'dark'); } catch (e) {}
    paint();
  });
})();

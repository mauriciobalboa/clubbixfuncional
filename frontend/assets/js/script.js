/* ========================================================================
   CLUBBIX — interatividade v5
   tab panels animados · reveals · contadores · calc cursos · indicação · Bixie
   ======================================================================== */
(function () {
  'use strict';

  var $ = function (s, c) { return (c || document).querySelector(s); };
  var $$ = function (s, c) { return Array.prototype.slice.call((c || document).querySelectorAll(s)); };
  var reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  var WHATS   = 'https://wa.me/5565992527948';

  /* =========================================================================
     UTILITÁRIOS DE SEGURANÇA — debounce · throttle · sanitize · btnBusy
     ========================================================================= */
  function debounce(fn, wait) {
    var timer;
    return function () {
      var ctx = this, args = arguments;
      clearTimeout(timer);
      timer = setTimeout(function () { fn.apply(ctx, args); }, wait);
    };
  }
  function throttle(fn, limit) {
    var last = 0;
    return function () {
      var now = Date.now();
      if (now - last < limit) return;
      last = now;
      return fn.apply(this, arguments);
    };
  }
  function sanitizeText(s) {
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#x27;');
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

  /* ─── ano no rodapé ─────────────────────────────────────────────────── */
  var yearEl = $('#year');
  if (yearEl) yearEl.textContent = new Date().getFullYear();

  /* =========================================================================
     NAV DINÂMICO — adapta botões conforme sessão ativa
     ========================================================================= */
  (function () {
    var navLoginBtn   = $('#navLoginBtn');
    var navAccountBtn = $('#navAccountBtn');
    var navCTABtn     = $('#navCTABtn');
    var navSess     = null;
    try { navSess = JSON.parse(localStorage.getItem('clubbix_session') || sessionStorage.getItem('clubbix_session')); } catch (e) {}

    if (!navSess) return; /* sem sessão: botões padrão ficam como estão */

    /* ── mostra "Minha conta" ── */
    if (navAccountBtn) navAccountBtn.hidden = false;

    /* ── "Entrar" → "Sair" ── */
    if (navLoginBtn) {
      navLoginBtn.textContent = 'Sair';
      navLoginBtn.removeAttribute('href');
      navLoginBtn.style.cursor = 'pointer';
      navLoginBtn.addEventListener('click', function (e) {
        e.preventDefault();
        localStorage.removeItem('clubbix_session');
        sessionStorage.removeItem('clubbix_session');
        window.location.reload();
      });
    }

    /* ── "Fazer parte" → "Upgrade" ou badge de máximo ── */
    if (navCTABtn) {
      var plano = (navSess.plano || '').toLowerCase();
      var isBest = plano.indexOf('max') !== -1 || plano.indexOf('fim') !== -1;
      if (isBest) {
        navCTABtn.textContent = '✦ No máximo!';
        navCTABtn.removeAttribute('href');
        navCTABtn.classList.add('nav-max-badge');
        navCTABtn.title = 'Você já está aproveitando o Clubbix ao máximo!';
        navCTABtn.style.cursor = 'default';
      } else {
        navCTABtn.textContent = 'Upgrade de plano';
        navCTABtn.href = 'assinar.html';
      }
    }
  })();

  /* ─── header ao rolar ────────────────────────────────────────────────── */
  var header  = $('.site-header');
  var toTop   = $('#toTop');
  var ticking = false;
  function onScroll() {
    var y = window.pageYOffset;
    if (header) header.classList.toggle('scrolled', y > 8);
    if (toTop)  toTop.classList.toggle('show', y > 520);
    ticking = false;
  }
  window.addEventListener('scroll', function () {
    if (!ticking) { window.requestAnimationFrame(onScroll); ticking = true; }
  }, { passive: true });
  onScroll();
  if (toTop) toTop.addEventListener('click', function () {
    window.scrollTo({ top: 0, behavior: reduced ? 'auto' : 'smooth' });
  });

  /* ─── reveals ao surgir (todos os tipos) ─────────────────────────────── */
  function registerReveals() {
    var allReveals = $$('.reveal, .reveal-left, .reveal-right, .reveal-scale');
    allReveals.forEach(function (el) {
      if (el._revealRegistered) return;
      el._revealRegistered = true;
      var parent = el.parentElement;
      var sibs = Array.prototype.filter.call(parent.children, function (c) {
        return c.classList.contains('reveal') ||
               c.classList.contains('reveal-left') ||
               c.classList.contains('reveal-right') ||
               c.classList.contains('reveal-scale');
      });
      var idx = sibs.indexOf(el);
      if (idx > 0) el.style.transitionDelay = Math.min(idx, 5) * 0.09 + 's';
    });
    if ('IntersectionObserver' in window && !reduced) {
      var obs = new IntersectionObserver(function (entries) {
        entries.forEach(function (e) {
          if (e.isIntersecting) { e.target.classList.add('is-visible'); obs.unobserve(e.target); }
        });
      }, { threshold: 0.1, rootMargin: '0px 0px -6% 0px' });
      allReveals.forEach(function (el) { if (!el.classList.contains('is-visible')) obs.observe(el); });
    } else {
      allReveals.forEach(function (el) { el.classList.add('is-visible'); });
    }
  }
  registerReveals();

  /* ─── contadores ─────────────────────────────────────────────────────── */
  function animateNumber(el, target, opts) {
    opts = opts || {};
    var prefix = opts.prefix || '', suffix = opts.suffix || '';
    if (reduced) { el.textContent = prefix + target + suffix; return; }
    var dur = 1400, start = null;
    function step(ts) {
      if (!start) start = ts;
      var p = Math.min((ts - start) / dur, 1);
      el.textContent = prefix + Math.round(target * (1 - Math.pow(1 - p, 3))) + suffix;
      if (p < 1) window.requestAnimationFrame(step);
      else el.textContent = prefix + target + suffix;
    }
    window.requestAnimationFrame(step);
  }
  function registerCounters() {
    var counters = $$('.counter');
    if (!('IntersectionObserver' in window)) {
      counters.forEach(function (el) { el.textContent = (el.dataset.prefix || '') + el.dataset.target; });
      return;
    }
    var cObs = new IntersectionObserver(function (entries) {
      entries.forEach(function (e) {
        if (e.isIntersecting) {
          animateNumber(e.target, parseInt(e.target.dataset.target, 10) || 0, { prefix: e.target.dataset.prefix });
          cObs.unobserve(e.target);
        }
      });
    }, { threshold: 0.5 });
    counters.forEach(function (el) { cObs.observe(el); });
  }
  registerCounters();

  /* ─── barra de progresso ─────────────────────────────────────────────── */
  function registerProgress() {
    var progress = $('.progress');
    if (!progress) return;
    if (!('IntersectionObserver' in window)) {
      var b = $('.progress-bar', progress);
      if (b) b.style.width = (b.dataset.fill || 0) + '%';
      return;
    }
    var pObs = new IntersectionObserver(function (entries) {
      entries.forEach(function (e) {
        if (!e.isIntersecting) return;
        var bar = $('.progress-bar', progress);
        var pct = $('.progress-pct', progress);
        var fill = parseInt(bar ? bar.dataset.fill : 0, 10) || 0;
        setTimeout(function () { if (bar) bar.style.width = fill + '%'; }, 200);
        if (pct) animateNumber(pct, parseInt(pct.dataset.target, 10) || fill, { suffix: '%' });
        pObs.unobserve(progress);
      });
    }, { threshold: 0.4 });
    pObs.observe(progress);
  }
  registerProgress();

  /* =========================================================================
     CLONAGEM DO HERO PARA TODAS AS ABAS
     ========================================================================= */
  var heroSection = $('#inicio');
  if (heroSection) {
    $$('.tab-panel').forEach(function(panel, idx) {
      if (idx === 0) return;
      var clone = heroSection.cloneNode(true);
      clone.removeAttribute('id');
      
      // Ajusta o botão "Conhecer os planos" para trocar de aba e rolar
      var anchorLinks = clone.querySelectorAll('a[href^="#planos-content"]');
      anchorLinks.forEach(function(link) {
        link.addEventListener('click', function(e) {
          e.preventDefault();
          switchPanel(0);
          setTimeout(function() {
            var target = $('#planos-content');
            if (target) {
              target.scrollIntoView({ behavior: 'smooth' });
            }
          }, 50);
        });
      });
      
      var inner = panel.querySelector('.panel-inner');
      if (inner) inner.insertBefore(clone, inner.firstChild);
    });
  }

  /* =========================================================================
     SISTEMA DE TAB PANELS COM ANIMAÇÃO pageEnterRight / pageEnterLeft
     ========================================================================= */
  var tabBtns      = $$('.tab[data-panel]');
  var panels       = $$('.tab-panel');
  var activePanel  = 0;
  var transitioning = false;
  var EXIT_DUR  = 600;  // ms — 0.6s slide exit
  var ENTER_DUR = 600;  // ms — 0.6s slide enter

  function scrollToPanels(targetPanel, smooth) {
    if (!targetPanel) targetPanel = document.querySelector('.tab-panel.is-active');
    if (!targetPanel) return;
    
    var contentStart = targetPanel.querySelector('.panel-inner > *:not(.hero)');
    if (!contentStart) contentStart = document.getElementById('tab-root');
    if (!contentStart) return;
    
    var offset = contentStart.getBoundingClientRect().top + window.pageYOffset - 90;
    
    if (smooth) {
      window.scrollTo({ top: offset, behavior: 'smooth' });
    } else {
      /* Desativa temporariamente o smooth scroll do CSS para forçar pulo imediato */
      var html = document.documentElement;
      var oldBehavior = html.style.scrollBehavior;
      html.style.scrollBehavior = 'auto';
      
      window.scrollTo(0, offset);
      
      /* Restaura o CSS original no próximo tick */
      setTimeout(function() {
        html.style.scrollBehavior = oldBehavior;
      }, 10);
    }
  }

  function updateNavIndicator() {
    var nav = $('#mainNav');
    var indicator = $('#navIndicator');
    if (!nav || !indicator) return;
    var activeTab = nav.querySelector('.tab.is-active');
    if (!activeTab) return;
    
    var navRect = nav.getBoundingClientRect();
    var tabRect = activeTab.getBoundingClientRect();
    
    var left = tabRect.left - navRect.left;
    indicator.style.transform = 'translateX(' + left + 'px)';
    indicator.style.width = tabRect.width + 'px';
  }

  window.addEventListener('resize', updateNavIndicator);
  window.addEventListener('load', updateNavIndicator);

  function setTabActive(idx) {
    tabBtns.forEach(function (btn, i) {
      btn.classList.toggle('is-active', i === idx);
      btn.setAttribute('aria-selected', i === idx ? 'true' : 'false');
    });
    // Request animation frame ensures DOM is updated before measuring
    requestAnimationFrame(updateNavIndicator);
  }
  
  // Call initially to set up
  updateNavIndicator();

  function switchPanel(newIdx) {
    if (transitioning) return;
    
    if (newIdx === activePanel) {
      scrollToPanels(panels[newIdx], true);
      return;
    }
    
    transitioning = true;

    var direction = newIdx > activePanel ? 'next' : 'prev';
    var prevPanel = panels[activePanel];
    var nextPanel = panels[newIdx];

    setTabActive(newIdx);

    if (reduced) {
      /* sem animação — troca direta */
      prevPanel.classList.remove('is-active');
      activePanel = newIdx;
      nextPanel.classList.add('is-active');
      registerReveals();
      registerCounters();
      registerProgress();
      transitioning = false;
      scrollToPanels();
      return;
    }

    /* ── FASE 1: anima saída do painel atual e entrada do novo simultaneamente ── */
    var exitClass = direction === 'next' ? 'exit-next' : 'exit-prev';
    var enterClass = direction === 'next' ? 'enter-next' : 'enter-prev';
    
    prevPanel.classList.add('is-exiting', exitClass);
    
    activePanel = newIdx;
    nextPanel.classList.add('is-active', enterClass);

    /* re-registra reveals e contadores do novo painel */
    registerReveals();
    registerCounters();
    registerProgress();

    /* ── FASE 2: limpa o painel antigo após a sua saída (340ms) ── */
    setTimeout(function () {
      prevPanel.classList.remove('is-active', 'is-exiting', exitClass);
    }, EXIT_DUR);

    /* ── FASE 3: limpa classes do novo painel após entrada terminar (620ms) ── */
    setTimeout(function () {
      nextPanel.classList.remove(enterClass);
      transitioning = false;
    }, ENTER_DUR);

    scrollToPanels(nextPanel);
  }

  /* Clique nos botões de tab do header */
  tabBtns.forEach(function (btn, i) {
    btn.addEventListener('click', function () { switchPanel(i); });
  });

  /* Links externos que abrem um painel específico (rodapé, hero CTA) */
  $$('[data-panel-link]').forEach(function (el) {
    el.addEventListener('click', function (e) {
      e.preventDefault();
      var idx = parseInt(el.getAttribute('data-panel-link'), 10);
      switchPanel(idx);
    });
  });

  /* Inicializa estados de aria nos botões */
  setTabActive(0);

  /* =========================================================================
     CALCULADORA DE CURSOS — dados reais UNIVAG 2026/1 (f=full, w=with convênio)
     ========================================================================= */
  var PERIOD_LABELS = {
    'M': 'Manhã', 'N': 'Noturno', 'I': 'Integral',
    'MN': 'Manhã / Noturno', 'SEMIP': 'Semipresencial',
    'EAD_DIG': 'EAD Digital', 'EAD_VIVO': 'EAD ao Vivo'
  };
  var COURSES = {
    /* Bacharelados / Licenciaturas */
    adm:     { p:{ M:{f:1690.50,w:929.78},  N:{f:1690.50,w:676.20},  EAD_DIG:{f:598.50,w:269.33},  EAD_VIVO:{f:1690.50,w:676.20}  }},
    agro:    { p:{ I:{f:2551.50,w:1658.48}, N:{f:2294.25,w:1032.41}, SEMIP:{f:1097.00,w:658.20}                                     }},
    arq:     { p:{ M:{f:2814.00,w:1547.70}, N:{f:2814.00,w:844.20},  SEMIP:{f:955.50,w:621.08}                                      }},
    bio:     { p:{ M:{f:2677.50,w:1472.63}, N:{f:2677.50,w:1124.55}, SEMIP:{f:1083.60,w:595.98}                                     }},
    cont:    { p:{ M:{f:1680.00,w:1008.00}, N:{f:1680.00,w:672.00},  EAD_DIG:{f:598.50,w:269.33},  EAD_VIVO:{f:1680.00,w:672.00}  }},
    chumanas:{ p:{ SEMIP:{f:530.25,w:291.64}                                                                                         }},
    comso:   { p:{ M:{f:1774.50,w:1064.70}, N:{f:1774.50,w:887.25},  EAD_DIG:{f:677.25,w:338.63},  EAD_VIVO:{f:1774.50,w:887.25}  }},
    dir:     { p:{ M:{f:1911.00,w:1242.15}, N:{f:1911.00,w:955.50}                                                                  }},
    edfis:   { p:{ M:{f:1501.50,w:975.98},  N:{f:1501.50,w:750.75},  SEMIP:{f:687.75,w:412.65}                                     }},
    enf:     { p:{ M:{f:1764.00,w:970.20},  N:{f:1764.00,w:882.00}                                                                  }},
    engamb:  { p:{ SEMIP:{f:955.50,w:573.30}                                                                                        }},
    engciv:  { p:{ M:{f:2073.75,w:1244.25}, N:{f:2073.75,w:933.19},  SEMIP:{f:955.50,w:573.30}                                     }},
    engali:  { p:{ SEMIP:{f:955.50,w:573.30}                                                                                        }},
    engprod: { p:{ N:{f:2073.75,w:933.19},  SEMIP:{f:955.50,w:573.30}                                                               }},
    engsw:   { p:{ M:{f:1989.75,w:1094.36}, N:{f:1989.75,w:994.88},  EAD_DIG:{f:677.25,w:372.49},  EAD_VIVO:{f:1989.75,w:994.88} }},
    engele:  { p:{ N:{f:2073.75,w:933.19},  SEMIP:{f:955.50,w:573.30}                                                               }},
    farm:    { p:{ M:{f:2903.25,w:1306.46}, N:{f:2903.25,w:1016.14}, SEMIP:{f:1197.00,w:538.65}                                    }},
    fisio:   { p:{ M:{f:2698.50,w:1430.21}, N:{f:2698.50,w:1106.39}, SEMIP:{f:1197.00,w:598.50}                                    }},
    fono:    { p:{ M:{f:2404.50,w:1322.48}, N:{f:2404.50,w:1009.89}, SEMIP:{f:1083.60,w:541.80}                                    }},
    let:     { p:{ N:{f:1228.50,w:491.40},  SEMIP:{f:530.25,w:291.64}                                                               }},
    matfis:  { p:{ SEMIP:{f:530.25,w:291.64}                                                                                        }},
    nutri:   { p:{ M:{f:2436.00,w:1339.80}, N:{f:2436.00,w:1023.12}, SEMIP:{f:1083.60,w:541.80}                                    }},
    odonto:  { p:{ I:{f:4399.50,w:2859.68}, N:{f:3517.50,w:2110.50}                                                                 }},
    ped:     { p:{ N:{f:1228.50,w:491.40},  SEMIP:{f:530.25,w:291.64}                                                               }},
    psico:   { p:{ M:{f:2231.25,w:1227.19}, N:{f:2231.25,w:1115.63}                                                                 }},
    servsoc: { p:{ SEMIP:{f:619.50,w:278.78}                                                                                        }},
    si:      { p:{ N:{f:1989.75,w:994.88},  EAD_DIG:{f:677.25,w:372.49},  EAD_VIVO:{f:1989.75,w:994.88}                           }},
    terap:   { p:{ M:{f:2084.25,w:937.91},  N:{f:2084.25,w:833.70},  SEMIP:{f:1083.60,w:541.80}                                    }},
    /* CST — Graduação Tecnológica */
    ads:     { p:{ MN:{f:1632.75,w:898.01}, EAD_DIG:{f:598.50,w:299.25}, EAD_VIVO:{f:1632.75,w:898.01}                            }},
    bdd:     { p:{ MN:{f:1632.75,w:898.01}, EAD_DIG:{f:598.50,w:299.25}                                                            }},
    cdt:     { p:{ MN:{f:1648.50,w:906.68}, EAD_DIG:{f:598.50,w:299.25}, EAD_VIVO:{f:1632.75,w:898.01}                            }},
    comext:  { p:{ EAD_DIG:{f:598.50,w:257.36}, EAD_VIVO:{f:1648.50,w:659.40}                                                      }},
    desin:   { p:{ MN:{f:1648.50,w:989.10}, SEMIP:{f:1083.60,w:541.80}, EAD_DIG:{f:903.00,w:406.35}, EAD_VIVO:{f:1648.50,w:741.83}}},
    estet:   { p:{ M:{f:1648.50,w:906.68},  N:{f:1648.50,w:741.83},  SEMIP:{f:1013.25,w:557.29}                                   }},
    gastro:  { p:{ M:{f:2955.75,w:1330.09}, SEMIP:{f:1013.25,w:709.28}, EAD_DIG:{f:1013.25,w:557.29}, EAD_VIVO:{f:2955.75,w:1034.51}}},
    gcom:    { p:{ N:{f:1648.50,w:659.40},  EAD_DIG:{f:598.50,w:257.36}, EAD_VIVO:{f:1648.50,w:659.40}                            }},
    grh:     { p:{ N:{f:1648.50,w:659.40},  EAD_DIG:{f:598.50,w:257.36}, EAD_VIVO:{f:1648.50,w:659.40}                            }},
    gfin:    { p:{ N:{f:1648.50,w:659.40},  EAD_DIG:{f:598.50,w:257.36}, EAD_VIVO:{f:1648.50,w:659.40}                            }},
    gpub:    { p:{ N:{f:1648.50,w:659.40},  EAD_DIG:{f:598.50,w:257.36}, EAD_VIVO:{f:1648.50,w:659.40}                            }},
    ia:      { p:{ N:{f:1632.75,w:898.01},  EAD_DIG:{f:598.50,w:299.25}, EAD_VIVO:{f:1989.75,w:895.39}                            }},
    jog:     { p:{ EAD_DIG:{f:598.50,w:299.25}                                                                                      }},
    log:     { p:{ N:{f:1648.50,w:659.40},  EAD_DIG:{f:598.50,w:257.36}, EAD_VIVO:{f:1648.50,w:659.40}                            }},
    mkt:     { p:{ N:{f:1648.50,w:659.40},  EAD_DIG:{f:598.50,w:257.36}, EAD_VIVO:{f:1648.50,w:659.40}                            }},
    pgr:     { p:{ N:{f:1648.50,w:659.40},  EAD_DIG:{f:598.50,w:257.36}, EAD_VIVO:{f:1648.50,w:659.40}                            }},
    prodg:   { p:{ N:{f:1811.25,w:724.50},  SEMIP:{f:845.25,w:380.36}                                                               }},
    radio:   { p:{ M:{f:1648.50,w:906.68},  N:{f:1648.50,w:741.83},  SEMIP:{f:1083.60,w:541.80}                                   }},
    seginfo: { p:{ MN:{f:1632.75,w:898.01}                                                                                          }}
  };
  var PLAN_COST = 409.99;

  function brl(n) {
    return 'R$ ' + n.toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  }

  var courseSelect = $('#courseSelect');
  var periodSelect = $('#periodSelect');
  var periodWrap   = $('#periodWrap');
  var ccFull    = $('#ccFull');
  var ccWith    = $('#ccWith');
  var ccMonth   = $('#ccMonth');
  var ccYear    = $('#ccYear');
  var ccPayback = $('#ccPayback');
  var ccRight   = $('#ccRight');

  function resetCalc() {
    [ccFull, ccWith, ccMonth, ccYear, ccPayback].forEach(function(el) { if (el) el.textContent = '—'; });
    if (ccRight) ccRight.classList.add('is-empty');
  }

  function updateCourseCalc() {
    var cd = COURSES[courseSelect ? courseSelect.value : ''];
    var pk = periodSelect ? periodSelect.value : '';
    if (!cd || !pk || !cd.p[pk]) { resetCalc(); return; }
    var d = cd.p[pk];
    if (ccRight) ccRight.classList.remove('is-empty');
    var sav = d.f - d.w;
    var pb  = sav > 0 ? PLAN_COST / sav : null;
    var pbStr = !pb ? '—'
      : pb <= 1 ? 'menos de 1 mês!'
      : (Math.round(pb * 10) / 10).toFixed(1).replace('.', ',') + ' meses';
    if (ccFull)    ccFull.textContent    = brl(d.f);
    if (ccWith)    ccWith.textContent    = brl(d.w);
    if (ccMonth)   ccMonth.textContent   = brl(sav);
    if (ccYear)    ccYear.textContent    = brl(sav * 12);
    if (ccPayback) ccPayback.textContent = pbStr;
  }

  function updatePeriodSelect() {
    resetCalc();
    var cd = COURSES[courseSelect ? courseSelect.value : ''];
    if (!cd) { if (periodWrap) periodWrap.hidden = true; return; }
    var keys = Object.keys(cd.p);
    if (keys.length === 1) {
      if (periodWrap) periodWrap.hidden = true;
      if (periodSelect) periodSelect.value = keys[0];
      updateCourseCalc();
      return;
    }
    if (periodSelect) {
      periodSelect.innerHTML = '<option value="">— escolha o período —</option>';
      keys.forEach(function(k) {
        var o = document.createElement('option');
        o.value = k; o.textContent = PERIOD_LABELS[k] || k;
        periodSelect.appendChild(o);
      });
    }
    if (periodWrap) periodWrap.hidden = false;
  }

  if (courseSelect) courseSelect.addEventListener('change', updatePeriodSelect);
  if (periodSelect) periodSelect.addEventListener('change', updateCourseCalc);

  /* =========================================================================
     INDIQUE E GANHE — gerador de link wa.me por WhatsApp
     ========================================================================= */
  var CLUBBIX_NUM  = '5565992527948';
  var refBtn       = $('#refBtn');
  var refNameEl    = $('#refName');
  var refNameField = $('#refNameField');
  var refResult    = $('#refResult');
  var refLinkEl    = $('#refLink');
  var refCopyBtn   = $('#refCopy');
  var refCopyLbl   = $('#refCopyLabel');
  var refWhats     = $('#refWhatsBtn');
  var refError     = $('#refError');
  var refAuthGate  = $('#refAuthGate');
  var refMemberBadge = $('#refMemberBadge');
  var refCardLead  = $('#refCardLead');

  /* lê a sessão (mesmo key do painel.js) */
  var refSession = null;
  try { refSession = JSON.parse(localStorage.getItem('clubbix_session') || sessionStorage.getItem('clubbix_session')); } catch (e) {}

  function initialsRef(n) {
    var p = (n || '?').trim().split(/\s+/);
    return ((p[0]||'?')[0] + (p.length > 1 ? p[p.length-1][0] : '')).toUpperCase();
  }

  /* formata telefone para exibição: (65) 99999-0000 */
  function maskPhoneRef(v) {
    v = v.replace(/\D/g, '').slice(0, 11);
    if (v.length > 10) return v.replace(/(\d{2})(\d{5})(\d{1,4})/, '($1) $2-$3');
    if (v.length > 6)  return v.replace(/(\d{2})(\d{4,5})(\d{1,4})/, '($1) $2-$3');
    if (v.length > 2)  return v.replace(/(\d{2})(\d{1,5})/, '($1) $2');
    if (v.length > 0)  return '(' + v;
    return v;
  }

  /* aplica máscara enquanto digita */
  if (refNameEl) {
    refNameEl.addEventListener('input', function () {
      var p = refNameEl.selectionStart, len = refNameEl.value.length;
      refNameEl.value = maskPhoneRef(refNameEl.value);
      if (p === len) refNameEl.setSelectionRange(refNameEl.value.length, refNameEl.value.length);
    });
  }

  /* gate desativado; campo de telefone sempre visível */
  if (refAuthGate) refAuthGate.hidden = true;
  if (refBtn)      refBtn.hidden      = false;
  if (refNameField) refNameField.hidden = false;

  if (refSession) {
    /* logado: mostra badge com dados do membro */
    if (refMemberBadge) {
      refMemberBadge.hidden = false;
      var miEl = $('#refMemberInitials');
      var mnEl = $('#refMemberName');
      var mpEl = $('#refMemberPlan');
      if (miEl) miEl.textContent = initialsRef(refSession.nome);
      if (mnEl) mnEl.textContent = refSession.nome || 'Membro';
      if (mpEl) mpEl.textContent = refSession.plano || 'Clubbix';
    }
    if (refCardLead) refCardLead.textContent = 'Informe seu WhatsApp para gerar o link e receber R$ 50,00 via PIX por cada indicação.';
  } else {
    if (refMemberBadge) refMemberBadge.hidden = true;
    if (refCardLead) {
      refCardLead.hidden = false;
      refCardLead.textContent = 'Informe seu WhatsApp para criar seu link de desconto personalizado.';
    }
  }

  if (refBtn) {
    var _refHandler = debounce(function () {
      var raw    = refNameEl ? refNameEl.value.trim() : '';
      var digits = raw.replace(/\D/g, '');
      if (digits.length < 10) {
        if (refError) refError.textContent = 'Informe um WhatsApp válido com DDD (ex: 65 99999-0000). 🙂';
        if (refNameEl) refNameEl.focus();
        btnFree(refBtn);
        return;
      }
      if (refError) refError.textContent = '';

      var msgIndicado = 'Quero assinar com desconto. Fui indicado pelo Whats ' + digits;
      var url = 'https://wa.me/' + CLUBBIX_NUM + '?text=' + encodeURIComponent(msgIndicado);

      if (refLinkEl) refLinkEl.value = url;
      if (refResult) refResult.classList.add('visible');

      var walletName = $('#refWalletName');
      var walletCode = $('#refWalletCode');
      if (walletName) walletName.textContent = maskPhoneRef(digits);
      if (walletCode) walletCode.textContent = digits;

      var shareMsg = 'Oi! Tenho um link especial com R$ 10,00 de desconto no Clubbix para você. Clica aqui e já cai direto no WhatsApp deles: ' + url;
      if (refWhats) refWhats.href = 'https://wa.me/?text=' + encodeURIComponent(shareMsg);

      if (refResult && !reduced) setTimeout(function () {
        refResult.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
      }, 100);

      setTimeout(function () { btnFree(refBtn); }, 1000);
    }, 600);

    refBtn.addEventListener('click', function () {
      btnBusy(refBtn, '<svg class="ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><use href="#i-link"/></svg> Gerando...');
      _refHandler();
    });
  }

  if (refCopyBtn) {
    refCopyBtn.addEventListener('click', function () {
      var url = refLinkEl ? refLinkEl.value : '';
      if (!url) return;
      function done() {
        if (refCopyLbl) refCopyLbl.textContent = 'Copiado!';
        setTimeout(function () { if (refCopyLbl) refCopyLbl.textContent = 'Copiar'; }, 2200);
      }
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(url).then(done);
      } else {
        if (refLinkEl) { refLinkEl.select(); document.execCommand('copy'); }
        done();
      }
    });
  }

  /* ─── formulário de contato ─────────────────────────────────────────── */
  var form = $('#contactForm');
  if (form) {
    var _contactBtn = form.querySelector('[type="submit"]');
    var _contactHandler = throttle(function () {
      var fb      = $('#formFeedback');
      var nome    = ($('#cName') || {}).value || '';
      var assunto = ($('#cSubject') || {}).value || '';
      var msg     = ($('#cMsg') || {}).value || '';
      nome = nome.trim().slice(0, 100); msg = msg.trim().slice(0, 1000);
      if (!nome || !msg) {
        if (fb) fb.textContent = 'Preencha seu nome e a mensagem. 🙂';
        btnFree(_contactBtn);
        return;
      }
      var texto = 'Olá! Vim pelo site do Clubbix.\n\nNome: ' + nome + '\nAssunto: ' + assunto + '\nMensagem: ' + msg;
      if (fb) fb.textContent = 'Abrindo o WhatsApp... 💬';
      window.open(WHATS + '?text=' + encodeURIComponent(texto), '_blank', 'noopener noreferrer');
      form.reset();
      setTimeout(function () { btnFree(_contactBtn); if (fb) fb.textContent = ''; }, 5000);
    }, 5000);

    form.addEventListener('submit', function (ev) {
      ev.preventDefault();
      btnBusy(_contactBtn, '<svg class="ico"><use href="#i-whats"/></svg> Enviando...');
      _contactHandler();
    });
  }

  /* =========================================================================
     CHATBOT — BIXIE
     ========================================================================= */
  var FAQ = [
    { id:'oque',    chip:'O que é o Clubbix?',
      keys:['o que e','que e o clubbix','que e clubbix','como funciona','oque e','sobre o clubbix'],
      a:'O <strong>Clubbix</strong> é um clube exclusivo! 🎉 Com uma assinatura você acessa benefícios em <strong>+400 estabelecimentos</strong> selecionados em todo o Brasil — e quem estuda na UNIVAG tem benefícios adicionais na graduação.' },
    { id:'planos',  chip:'Planos e valores',
      keys:['plano','preco','preço','valor','quanto custa','custa','pagar','assinar'],
      a:'Temos 3 planos:<br>• <strong>Básico</strong> — R$ 409,99<br>• <strong>Plus</strong> — R$ 699,99<br>• <strong>Max</strong> — R$ 1.199,88 ★<br><a href="assinar.html">Ver planos »</a>' },
    { id:'acesso',  chip:'Como funciona o acesso?',
      keys:['como funciona','acesso','cartao','cartão','usar','estabelecimento','parceiro'],
      a:'Simples: assine o plano, receba seu <strong>cartão de membro digital</strong> e apresente em qualquer estabelecimento parceiro. O desconto é aplicado na hora. 📱' },
    { id:'univag',  chip:'Benefício na graduação',
      keys:['univag','graduacao','mensalidade','faculdade','curso','tecnologo'],
      a:'Membros com acesso à UNIVAG têm <strong>benefício adicional na mensalidade</strong> da sua graduação ou tecnólogo. Use a calculadora na aba Exclusividade para ver o seu curso. 📚' },
    { id:'medicina',chip:'Medicina tem benefício?',
      keys:['medicina','medico','médico'],
      a:'O curso de <strong>Medicina não participa</strong> do benefício na mensalidade. Mas você acessa todos os estabelecimentos parceiros do clube! 🙂' },
    { id:'indicar', chip:'Indique e Ganhe',
      keys:['indicar','indicacao','indicação','indique','link','amigo'],
      a:'No <strong>programa de indicação</strong> você gera seu link exclusivo e, quando um amigo aderir, vocês dois ganham vantagens! 🎁 <a href="#tab-root">Veja a aba Indique e Ganhe »</a>' },
    { id:'filant',  chip:'Filantropia',
      keys:['filantr','doacao','doação','ong','alimento','social'],
      a:'A <strong>Pegada Filantrópica</strong> ❤️ Parte de cada assinatura vira alimento doado para ONGs de Cuiabá. Ano passado entregamos 500 kg — este ano vamos entregar 1 tonelada e meia! Veja os números na aba Filantropia.' },
    { id:'login',   chip:'Acessar minha conta',
      keys:['acesso','login','entrar','minha conta','ja sou','já sou'],
      a:'Se você já é membro: <a href="login.html">Entrar na minha conta »</a> 📱' },
    { id:'contato', chip:'Falar com vocês',
      keys:['contato','telefone','email','e-mail','atendimento','falar'],
      a:'WhatsApp: <strong>(65) 99252-7948</strong> · E-mail: <strong>contato@clubbix.com.br</strong>. <a href="' + WHATS + '" target="_blank" rel="noopener">Abrir WhatsApp »</a>' }
  ];
  var MAIN_CHIPS = ['oque','planos','acesso','univag','indicar','whatsapp'];

  var launcher  = $('#chatLauncher');
  var chatbot   = $('#chatbot');
  var chatBody  = $('#chatBody');
  var chatForm2 = $('#chatForm');
  var chatInput = $('#chatInput');
  var chatClose = $('#chatClose');
  var chatStarted = false;

  function norm(s){ return s.toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g,''); }
  function scrollChat(){ if(chatBody) chatBody.scrollTop = chatBody.scrollHeight; }

  function addMsg(who, html) {
    var msg = document.createElement('div');
    msg.className = 'msg ' + who;
    if (who === 'bot') {
      var av = document.createElement('img');
      av.className = 'msg-avatar'; av.src = 'assets/img/logo.png'; av.alt = '';
      msg.appendChild(av);
    }
    var bub = document.createElement('div');
    bub.className = 'msg-bubble'; bub.innerHTML = html;
    msg.appendChild(bub);
    if (chatBody) chatBody.appendChild(msg);
    scrollChat();
  }

  function showTyping() {
    var t = document.createElement('div');
    t.className = 'msg bot'; t.id = 'typing';
    t.innerHTML = '<img class="msg-avatar" src="assets/img/logo.png" alt=""><div class="msg-bubble chat-typing"><span></span><span></span><span></span></div>';
    if (chatBody) chatBody.appendChild(t); scrollChat();
  }
  function hideTyping() { var t = $('#typing'); if (t && t.parentNode) t.parentNode.removeChild(t); }

  function renderChips(ids) {
    var old = chatBody ? $('.quick-replies', chatBody) : null;
    if (old && old.parentNode) old.parentNode.removeChild(old);
    var box = document.createElement('div'); box.className = 'quick-replies';
    ids.forEach(function (id) {
      var lbl, action;
      if (id === 'whatsapp') { lbl = '💬 WhatsApp'; action = 'whatsapp'; }
      else {
        var item = FAQ.filter(function(f){ return f.id===id; })[0];
        if (!item) return;
        lbl = item.chip; action = id;
      }
      var btn = document.createElement('button');
      btn.className = 'qr'; btn.type = 'button'; btn.textContent = lbl;
      btn.addEventListener('click', function () {
        if (action === 'whatsapp') { addMsg('user','Quero falar no WhatsApp'); botReply('whatsapp'); }
        else { var f=FAQ.filter(function(x){return x.id===action;})[0]; addMsg('user',f.chip); botReply(action); }
      });
      box.appendChild(btn);
    });
    if (chatBody) chatBody.appendChild(box); scrollChat();
  }

  function answerFor(text) {
    var n = norm(text);
    if (/\b(oi|ola|opa|eai|e ai|bom dia|boa tarde|boa noite|hey)\b/.test(n))
      return { a:'Oi! 😄 Sobre o que você quer saber?', chips: MAIN_CHIPS };
    if (/(obrigad|valeu|brigad)/.test(n))
      return { a:'Por nada! 🧡 Se precisar, é só chamar.', chips: MAIN_CHIPS };
    if (/(tchau|ate mais|falou)/.test(n))
      return { a:'Até logo! Aproveite o Clubbix. 👋', chips: MAIN_CHIPS };
    for (var i=0; i<FAQ.length; i++)
      for (var k=0; k<FAQ[i].keys.length; k++)
        if (n.indexOf(norm(FAQ[i].keys[k])) !== -1) return { a: FAQ[i].a, chips: MAIN_CHIPS };
    return { a:'Não tenho certeza sobre isso. 🤔 Posso ajudar com os tópicos abaixo ou te conectar ao nosso time no WhatsApp <strong>(65) 99252-7948</strong>.', chips:['oque','planos','indicar','filant','whatsapp'] };
  }

  function botReply(idOrText, isFreeText) {
    showTyping();
    setTimeout(function () {
      hideTyping();
      if (idOrText === 'whatsapp') {
        addMsg('bot','Abrindo o WhatsApp! Se não abrir, o número é <strong>(65) 99252-7948</strong>. 💬');
        window.open(WHATS,'_blank','noopener noreferrer'); renderChips(MAIN_CHIPS); return;
      }
      var res = isFreeText ? answerFor(idOrText) : (function(){
        var f=FAQ.filter(function(x){return x.id===idOrText;})[0];
        return { a: f?f.a:'...', chips: MAIN_CHIPS };
      })();
      addMsg('bot', res.a); renderChips(res.chips);
    }, reduced ? 180 : 650 + Math.random() * 400);
  }

  function startChat() {
    if (chatStarted) return; chatStarted = true;
    addMsg('bot','Oi! 👋 Eu sou o <strong>Bixie</strong>, assistente do Clubbix. Posso tirar suas dúvidas sobre planos, benefícios, indicações e mais.');
    setTimeout(function () {
      addMsg('bot','Escolha um assunto ou escreva sua pergunta. 👇');
      renderChips(['oque','planos','acesso','univag','indicar','filant','login','whatsapp']);
    }, reduced ? 150 : 560);
  }

  function openChat() {
    if (!chatbot) return;
    chatbot.classList.add('is-open'); chatbot.setAttribute('aria-hidden','false');
    if (launcher) launcher.classList.add('hidden');
    startChat();
    if (window.innerWidth > 640) setTimeout(function(){ if(chatInput) chatInput.focus(); }, 360);
  }
  function closeChat() {
    if (!chatbot) return;
    chatbot.classList.remove('is-open'); chatbot.setAttribute('aria-hidden','true');
    if (launcher) launcher.classList.remove('hidden');
  }

  if (launcher) launcher.addEventListener('click', openChat);
  if (chatClose) chatClose.addEventListener('click', closeChat);
  document.addEventListener('keydown', function (e) {
    if (e.key === 'Escape' && chatbot && chatbot.classList.contains('is-open')) closeChat();
  });
  $$('[data-open-chat]').forEach(function (el) {
    el.addEventListener('click', openChat);
    el.addEventListener('keydown', function (e) { if (e.key==='Enter'||e.key===' '){e.preventDefault();openChat();} });
  });
  if (chatForm2) chatForm2.addEventListener('submit', function (e) {
    e.preventDefault();
    var text = chatInput ? chatInput.value.trim() : '';
    if (!text) return;
    addMsg('user', sanitizeText(text));
    if (chatInput) chatInput.value = '';
    botReply(text, true);
  });

})();

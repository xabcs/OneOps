define('vs/nls.messages-loader', ['exports'], function (s) {
  function a(o, l, n, t) {
    const e = t['vs/nls']?.availableLanguages?.['*'];
    !e || e === 'en'
      ? n({})
      : l([`vs/nls.messages.${e}`], () => {
          n({});
        });
  }
  ((s.load = a), Object.defineProperty(s, Symbol.toStringTag, { value: 'Module' }));
});

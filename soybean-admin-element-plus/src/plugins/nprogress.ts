import NProgress from 'nprogress';

/** Setup plugin NProgress */
export function setupNProgress() {

  NProgress.configure({
    easing: 'ease',
    /** speed: 10,  // 减少速度 */
  });
  // mount on window
  window.NProgress = NProgress;
}

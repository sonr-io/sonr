export function isAndroid() {
  return navigator && /Android/i.test(navigator.userAgent);
}

/**
 * @see https://stackoverflow.com/a/58065241
 */
export function isIOS() {
  return (
    navigator &&
    (/iPhone|iPad|iPod/i.test(navigator.userAgent) ||
      (navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1))
  );
}

export function isMobile() {
  return isAndroid() || isIOS();
}

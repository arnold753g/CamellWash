export const useBranding = () => {
  const config = useRuntimeConfig()
  return {
    name: String(config.public.appName),
    shortName: String(config.public.appShortName),
    logo: '/branding/camell-wash-logo.png',
    favicon: '/branding/camell-wash-logo.png'
  }
}

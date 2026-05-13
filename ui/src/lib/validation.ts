import type { ConnectionProfile, SupportedProviderOption } from "@/lib/types";

export function validateProfile(
  profile: ConnectionProfile,
  apiKey: string,
  supportedProviders: SupportedProviderOption[],
  hasSavedKey = false,
) {
  const errors: string[] = [];
  const providerName = profile.provider.name.trim();
  if (!providerName) {
    errors.push("A compatible connection is required.");
    return errors;
  }

  const descriptor = supportedProviders.find((item) => item.name === "compatible");
  if (!descriptor) {
    errors.push(`Provider "${providerName}" is not supported by the local backend.`);
    return errors;
  }

  if (!profile.provider.model.trim()) {
    errors.push("Model is required.");
  }

  if (!apiKey.trim() && !hasSavedKey) {
    errors.push("API key is required.");
  }

  if (!profile.provider.baseUrl.trim()) {
    errors.push("Base URL is required.");
  } else {
    try {
      const url = new URL(profile.provider.baseUrl.trim());
      if (!["http:", "https:"].includes(url.protocol) || !url.hostname || url.username || url.password || url.search || url.hash) {
        errors.push("Base URL must be an HTTP(S) URL without credentials or query.");
      }
    } catch {
      errors.push("Base URL must be a valid HTTP(S) URL.");
    }
  }

  return errors;
}

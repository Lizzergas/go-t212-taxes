class T212Taxes < Formula
  desc "A comprehensive tool for processing Trading 212 CSV exports and calculating tax obligations"
  homepage "https://github.com/Lizzergas/go-t212-taxes"
  url "https://github.com/Lizzergas/go-t212-taxes/releases/download/v1.0.0/go-t212-taxes-darwin-x86_64.tar.gz"
  sha256 "87e71aeba3708c860e7393c47f349e479296999c2eb7223ef6a52fa348c53c5f"
  license "MIT"
  version "1.0.0"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/Lizzergas/go-t212-taxes/releases/download/v1.0.0/go-t212-taxes-darwin-x86_64.tar.gz"
      sha256 "87e71aeba3708c860e7393c47f349e479296999c2eb7223ef6a52fa348c53c5f"
    end
    if Hardware::CPU.arm?
      url "https://github.com/Lizzergas/go-t212-taxes/releases/download/v1.0.0/go-t212-taxes-darwin-arm64.tar.gz"
      sha256 "043468f27483a83e80179d9a68466f864dcf4e03293a24d3653fe2b9d55d00c4"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/Lizzergas/go-t212-taxes/releases/download/v1.0.0/go-t212-taxes-linux-x86_64.tar.gz"
      sha256 "4336408a1e69c45612ef5d28fd61f9e11c0b05865a7c4a82b778c43cf80fb8ab"
    end
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/Lizzergas/go-t212-taxes/releases/download/v1.0.0/go-t212-taxes-linux-arm64.tar.gz"
      sha256 "d9a1953702dea0e8dab7d1e9fb137f9a59ad756bdb4114e2bc6a8be5c9367841"
    end
  end

  def install
    bin.install "t212-taxes"
    etc.install "config.yaml" => "t212-taxes/config.yaml"
  end

  test do
    system "#{bin}/t212-taxes", "version"
  end
end 
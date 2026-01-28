import argparse
import asyncio
import os

from playwright.async_api import async_playwright

# 修复 Windows 下 Playwright 可能找不到 $HOME 的问题
if os.name == "nt" and "HOME" not in os.environ:
    os.environ["HOME"] = os.environ.get("USERPROFILE", "")


async def run(url, screenshot_path=None):
    async with async_playwright() as p:
        # 优先使用已安装的 chrome，或者默认的 chromium
        browser = await p.chromium.launch(headless=True)
        page = await browser.new_page()

        print(f"[*] 正在访问: {url}")
        try:
            await page.goto(url, wait_until="networkidle", timeout=60000)

            title = await page.title()
            print(f"[+] 页面标题: {title}")

            # 提取主要文本内容 (简单的概要)
            content = await page.inner_text("body")
            print(f"[+] 内容摘要 (前500字):\n{content[:500]}...")

            if screenshot_path:
                await page.screenshot(path=screenshot_path)
                print(f"[+] 截图已保存至: {screenshot_path}")

        except Exception as e:
            print(f"[!] 访问失败: {e}")
        finally:
            await browser.close()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="使用 Playwright 访问网页的脚本")
    parser.add_argument("--url", required=True, help="要访问的 URL")
    parser.add_argument("--screenshot", help="截图保存路径")

    args = parser.parse_args()

    try:
        asyncio.run(run(args.url, args.screenshot))
    except KeyboardInterrupt:
        pass
